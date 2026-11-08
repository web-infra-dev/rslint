// A test-only host: supplies a bound Program and an injected synchronous
// query endpoint. No linter entrypoint or Go rule participates in this host.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/typeservice"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type request struct {
	MetadataOnly bool                `json:"metadataOnly"`
	Method       string              `json:"method"`
	Config       string              `json:"config"`
	File         string              `json:"file"`
	Query        typeservice.Query   `json:"query"`
	Queries      []typeservice.Query `json:"queries"`
	Generation   uint32              `json:"generation"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 32<<20)
	output := json.NewEncoder(os.Stdout)
	var session *typeservice.Session
	var p *program.Program
	var projector estree.Builder
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	for scanner.Scan() {
		func() {
			var req request
			var value any
			var err error
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("%v", recovered)
				}
				response := map[string]any{"result": value}
				if err != nil {
					response = map[string]any{"error": err.Error()}
				}
				if err := output.Encode(response); err != nil {
					panic(err)
				}
			}()
			if err = json.Unmarshal(scanner.Bytes(), &req); err != nil {
				return
			}
			switch req.Method {
			case "init":
				if session != nil {
					session.Close()
				}
				cwd, readErr := os.Getwd()
				if readErr != nil {
					err = readErr
					return
				}
				cwd = tspath.NormalizePath(cwd)
				fs := bundled.WrapFS(osvfs.FS())
				host := utils.CreateCompilerHost(cwd, fs)
				raw, buildErr := utils.CreateProgram(false, fs, cwd, req.Config, host)
				if buildErr != nil {
					err = buildErr
					return
				}
				p = program.NewFromCompiler(raw)
				eligible := []*ast.SourceFile{}
				for _, name := range p.RootFileNames() {
					if f := p.GetSourceFile(name); f != nil {
						eligible = append(eligible, f)
					}
				}
				session, err = typeservice.New(context.Background(), p, eligible)
				if err == nil {
					value, err = session.Descriptor()
				}
			case "project":
				if session == nil {
					err = errors.New("session not initialized")
					return
				}
				f := p.GetSourceFile(tspath.NormalizePath(req.File))
				if f == nil {
					err = errors.New("missing source")
					return
				}
				binding, bindErr := session.BindTarget(f)
				if bindErr != nil {
					err = bindErr
					return
				}
				projection, projectErr := projector.Project(context.Background(), f, estree.Options{SourceType: "module", HasBOM: strings.HasPrefix(f.Text(), "\ufeff")})
				projector.Reset()
				value = struct {
					estree.Projection
					Binding typeservice.Binding `json:"binding"`
				}{projection, binding}
				err = projectErr
			case "source":
				f, sourceErr := session.Source(req.Generation, tspath.NormalizePath(req.File))
				if sourceErr != nil {
					err = sourceErr
					return
				}
				if f != nil {
					if req.MetadataOnly {
						value = struct {
							FileName string `json:"fileName"`
						}{f.FileName()}
						return
					}
					tree, _, encodeErr := encoder.EncodeSourceFile(f)
					err = encodeErr
					value = struct {
						FileName string `json:"fileName"`
						Tree     []byte `json:"tree"`
					}{f.FileName(), tree}
				}
			case "batch":
				if session == nil {
					err = errors.New("session not initialized")
					return
				}
				value, err = session.QueryBatch(context.Background(), req.Queries)
			case "query":
				if session == nil {
					err = errors.New("session not initialized")
					return
				}
				value, err = session.Query(context.Background(), req.Query)
			case "close":
				if session != nil {
					session.Close()
				}
				value = true
			default:
				err = errors.New("unsupported harness request")
			}
		}()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
