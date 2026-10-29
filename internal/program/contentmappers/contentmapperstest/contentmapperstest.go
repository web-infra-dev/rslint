// Package contentmapperstest provides an in-process content mapper for tests.
package contentmapperstest

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/contentmapper"
)

// Extension is the file extension the test mapper package handles.
const Extension = ".fake"

// Spawner is a contentmapper.Spawner that serves every spawn with an
// in-process mapper instead of a child process. The mapper turns each file
// into Output, as TypeScript.
type Spawner struct {
	Output string

	spawns atomic.Int32
	closes atomic.Int32
}

var _ contentmapper.Spawner = (*Spawner)(nil)

// Spawns returns how many mapper processes were started.
func (s *Spawner) Spawns() int { return int(s.spawns.Load()) }

// Closes returns how many mapper processes were stopped.
func (s *Spawner) Closes() int { return int(s.closes.Load()) }

func (s *Spawner) Spawn(command []string, dir string, stderr io.Writer) (io.ReadWriteCloser, error) {
	s.spawns.Add(1)
	client, server := net.Pipe()
	go serve(server, s.Output)
	return &process{Conn: client, closes: &s.closes}, nil
}

type process struct {
	net.Conn
	closes *atomic.Int32
	once   sync.Once
}

func (p *process) Close() error {
	p.once.Do(func() { p.closes.Add(1) })
	return p.Conn.Close()
}

// WritePackage writes a mapper package named name under dir/node_modules,
// declaring the manifest ts-go reads from package.json.
func WritePackage(t testing.TB, dir string, name string) {
	t.Helper()
	packageDir := filepath.Join(dir, "node_modules", filepath.FromSlash(name))
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf(`{"name":%q,"version":"1.0.0","typescript":{"contentMapper":{"exec":["node","server.js"]}}}`, name)
	if err := os.WriteFile(filepath.Join(packageDir, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

type message struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
}

// serve answers the mapper protocol over conn with Content-Length framing,
// like the LSP base protocol.
func serve(conn net.Conn, output string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		body, err := readFrame(reader)
		if err != nil {
			return
		}
		var request message
		if json.Unmarshal(body, &request) != nil || len(request.ID) == 0 {
			continue
		}
		var result any
		switch request.Method {
		case contentmapper.MethodInitialize:
			result = contentmapper.InitializeResult{PositionEncoding: contentmapper.PositionEncodingUTF8, DiagnosticSource: "fake"}
		case contentmapper.MethodOpenProject:
			// The manifest declares no dynamicConfig, so ts-go rejects a
			// config identity.
			result = contentmapper.OpenProjectResult{}
		case contentmapper.MethodTransform:
			result = contentmapper.TransformResult{MappedOutput: contentmapper.MappedOutput{Text: output, Extension: ".ts"}}
		}
		response, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
		if err != nil {
			return
		}
		if _, err := fmt.Fprintf(conn, "Content-Length: %d\r\n\r\n%s", len(response), response); err != nil {
			return
		}
	}
}

func readFrame(reader *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if value, ok := strings.CutPrefix(line, "Content-Length:"); ok {
			if length, err = strconv.Atoi(strings.TrimSpace(value)); err != nil {
				return nil, err
			}
		}
	}
	if length < 0 {
		return nil, errors.New("missing Content-Length")
	}
	body := make([]byte, length)
	_, err := io.ReadFull(reader, body)
	return body, err
}
