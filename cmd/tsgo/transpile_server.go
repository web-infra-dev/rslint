package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgpackArray3 = 0x93
	msgpackU8     = 0xcc
	msgpackBin8   = 0xc4
	msgpackBin16  = 0xc5
	msgpackBin32  = 0xc6

	messageTypeRequest  = 1
	messageTypeResponse = 4
	messageTypeError    = 5
)

type serverMessage struct {
	messageType byte
	method      string
	payload     []byte
}

type transpileServerRequest struct {
	FileName        string `json:"fileName"`
	ConfigFileName  string `json:"configFileName,omitempty"`
	Module          string `json:"module,omitempty"`
	Target          string `json:"target,omitempty"`
	Jsx             string `json:"jsx,omitempty"`
	InlineSourceMap bool   `json:"inlineSourceMap,omitempty"`
	SourceMap       bool   `json:"sourceMap,omitempty"`
	TypeCheck       bool   `json:"typecheck,omitempty"`
}

type transpileServerResponse struct {
	OutputText      string `json:"outputText,omitempty"`
	DiagnosticsText string `json:"diagnosticsText,omitempty"`
	HasError        bool   `json:"hasError"`
}

func runTranspileServer(cwd string) int {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		message, err := readServerMessage(reader)
		if err != nil {
			if err == io.EOF {
				return 0
			}
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
		if message.messageType != messageTypeRequest {
			_ = writeServerMessage(writer, messageTypeError, message.method, []byte("invalid request type"))
			continue
		}
		switch message.method {
		case "transpile":
			response, err := handleTranspileServerRequest(message.payload, cwd)
			if err != nil {
				_ = writeServerMessage(writer, messageTypeError, message.method, []byte(err.Error()))
				continue
			}
			payload, err := json.Marshal(response)
			if err != nil {
				_ = writeServerMessage(writer, messageTypeError, message.method, []byte(err.Error()))
				continue
			}
			if err := writeServerMessage(writer, messageTypeResponse, message.method, payload); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return 1
			}
		default:
			_ = writeServerMessage(writer, messageTypeError, message.method, []byte("unknown method"))
		}
	}
}

func handleTranspileServerRequest(payload []byte, cwd string) (*transpileServerResponse, error) {
	var request transpileServerRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("invalid request payload: %w", err)
	}
	if request.FileName == "" {
		return nil, fmt.Errorf("fileName is required")
	}
	opts := transpileOptions{
		Config:          request.ConfigFileName,
		File:            request.FileName,
		Module:          request.Module,
		Target:          request.Target,
		Jsx:             request.Jsx,
		InlineSourceMap: request.InlineSourceMap,
		SourceMap:       request.SourceMap,
		TypeCheck:       request.TypeCheck,
	}
	result, err := transpileForServer(opts, cwd)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func transpileForServer(opts transpileOptions, cwd string) (*transpileServerResponse, error) {
	if opts.File == "" {
		return nil, fmt.Errorf("file is required")
	}
	absFile, err := resolvePath(cwd, opts.File)
	if err != nil {
		return nil, fmt.Errorf("error resolving file path: %v", err)
	}
	absFile = tspath.NormalizePath(absFile)

	configPath := ""
	if opts.Config != "" {
		var err error
		configPath, err = resolvePath(cwd, opts.Config)
		if err != nil {
			return nil, fmt.Errorf("error resolving config path: %v", err)
		}
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	currentDirectory := tspath.NormalizePath(cwd)
	host := utils.CreateCompilerHost(currentDirectory, fs)

	var baseParsed *tsoptions.ParsedCommandLine
	if configPath != "" {
		parsed, diags := tsoptions.GetParsedCommandLineOfConfigFile(configPath, &core.CompilerOptions{}, host, nil)
		if len(diags) > 0 {
			return &transpileServerResponse{
				DiagnosticsText: formatTranspileDiagnostics(diags, cwd),
				HasError:        true,
			}, nil
		}
		if parsed == nil {
			return nil, fmt.Errorf("error parsing tsconfig: %s", configPath)
		}
		baseParsed = parsed
	}

	var compilerOptions core.CompilerOptions
	if baseParsed != nil && baseParsed.CompilerOptions() != nil {
		compilerOptions = *baseParsed.CompilerOptions()
	}
	if err := applyTranspileOverrides(&compilerOptions, opts, absFile); err != nil {
		return nil, err
	}

	comparePathsOptions := tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
		CurrentDirectory:          currentDirectory,
	}
	attachConfig := baseParsed != nil && shouldAttachConfig(baseParsed, absFile, comparePathsOptions)
	fileNames := []string{absFile}
	if attachConfig {
		fileNames = baseParsed.FileNames()
	}
	parsed := tsoptions.NewParsedCommandLine(&compilerOptions, fileNames, comparePathsOptions)
	if attachConfig {
		parsed.ConfigFile = baseParsed.ConfigFile
		parsed.Raw = baseParsed.Raw
		parsed.SetTypeAcquisition(baseParsed.TypeAcquisition())
	}

	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         parsed,
		SingleThreaded: core.TSTrue,
		Host:           host,
	})
	if program == nil {
		return nil, fmt.Errorf("error creating program")
	}
	sourceFile := program.GetSourceFile(absFile)
	if sourceFile == nil {
		return nil, fmt.Errorf("error loading source file: %s", absFile)
	}

	var diagnostics []*ast.Diagnostic
	if opts.TypeCheck {
		diagnostics = compiler.GetDiagnosticsOfAnyProgram(
			context.Background(),
			program,
			nil,
			false,
			program.GetBindDiagnostics,
			program.GetSemanticDiagnostics,
		)
	}

	var output string
	emitResult := program.Emit(context.Background(), compiler.EmitOptions{
		TargetSourceFile: sourceFile,
		EmitOnly:         compiler.EmitOnlyJs,
		WriteFile: func(fileName string, text string, writeByteOrderMark bool, data *compiler.WriteFileData) error {
			if isJsOutput(fileName) {
				output = text
			}
			return nil
		},
	})
	if emitResult != nil && len(emitResult.Diagnostics) > 0 {
		diagnostics = append(diagnostics, emitResult.Diagnostics...)
	}

	diagnosticsText := formatTranspileDiagnostics(diagnostics, cwd)
	hasError := hasErrorDiagnostics(diagnostics)
	if emitResult != nil && emitResult.EmitSkipped {
		hasError = true
		if diagnosticsText == "" {
			diagnosticsText = "emit skipped\n"
		} else {
			diagnosticsText += "emit skipped\n"
		}
	}
	if output == "" {
		hasError = true
		if diagnosticsText == "" {
			diagnosticsText = "no output produced\n"
		} else {
			diagnosticsText += "no output produced\n"
		}
	}

	return &transpileServerResponse{
		OutputText:      output,
		DiagnosticsText: diagnosticsText,
		HasError:        hasError,
	}, nil
}

func resolvePath(cwd string, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	return filepath.Join(cwd, value), nil
}

func readServerMessage(r *bufio.Reader) (*serverMessage, error) {
	header, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if header != msgpackArray3 {
		return nil, fmt.Errorf("invalid message header")
	}
	typeTag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if typeTag != msgpackU8 {
		return nil, fmt.Errorf("invalid message type tag")
	}
	messageType, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	methodBytes, err := readMsgpackBin(r)
	if err != nil {
		return nil, err
	}
	payloadBytes, err := readMsgpackBin(r)
	if err != nil {
		return nil, err
	}
	return &serverMessage{
		messageType: messageType,
		method:      string(methodBytes),
		payload:     payloadBytes,
	}, nil
}

func readMsgpackBin(r *bufio.Reader) ([]byte, error) {
	tag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	var size uint
	switch tag {
	case msgpackBin8:
		length, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		size = uint(length)
	case msgpackBin16:
		var length uint16
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		size = uint(length)
	case msgpackBin32:
		var length uint32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		size = uint(length)
	default:
		return nil, fmt.Errorf("invalid msgpack bin tag")
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, nil
}

func writeServerMessage(w *bufio.Writer, messageType byte, method string, payload []byte) error {
	if payload == nil {
		payload = []byte{}
	}
	if err := w.WriteByte(msgpackArray3); err != nil {
		return err
	}
	if err := w.WriteByte(msgpackU8); err != nil {
		return err
	}
	if err := w.WriteByte(messageType); err != nil {
		return err
	}
	if err := writeMsgpackBin(w, []byte(method)); err != nil {
		return err
	}
	if _, err := w.WriteString(method); err != nil {
		return err
	}
	if err := writeMsgpackBin(w, payload); err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return w.Flush()
}

func writeMsgpackBin(w *bufio.Writer, payload []byte) error {
	length := len(payload)
	switch {
	case length < 256:
		if err := w.WriteByte(msgpackBin8); err != nil {
			return err
		}
		return w.WriteByte(byte(length))
	case length < 1<<16:
		if err := w.WriteByte(msgpackBin16); err != nil {
			return err
		}
		return binary.Write(w, binary.BigEndian, uint16(length))
	default:
		if err := w.WriteByte(msgpackBin32); err != nil {
			return err
		}
		return binary.Write(w, binary.BigEndian, uint32(length))
	}
}
