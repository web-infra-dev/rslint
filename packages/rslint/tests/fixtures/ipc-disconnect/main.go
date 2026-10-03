// This peer exercises real stdout EOF independently of process exit on every
// platform. Node's Windows fs.close leaves standard descriptors 0-2 open.
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type message struct {
	Kind string          `json:"kind"`
	ID   int             `json:"id"`
	Data json.RawMessage `json:"data"`
}

func send(kind string, id int, data any) error {
	body, err := json.Marshal(struct {
		Kind string `json:"kind"`
		ID   int    `json:"id"`
		Data any    `json:"data"`
	}{kind, id, data})
	if err != nil {
		return err
	}
	frame := binary.LittleEndian.AppendUint32(nil, uint32(len(body)))
	_, err = os.Stdout.Write(append(frame, body...))
	return err
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected one disconnect mode")
	}
	mode := os.Args[1]
	switch mode {
	case "eof-before-init", "eof-after-init", "reject-init", "reject-init-eof":
	default:
		return fmt.Errorf("unknown disconnect mode %q", mode)
	}
	for {
		var size uint32
		if err := binary.Read(os.Stdin, binary.LittleEndian, &size); err != nil {
			return err
		}
		if size > 1024*1024 {
			return fmt.Errorf("unexpected fixture frame size %d", size)
		}
		body := make([]byte, size)
		if _, err := io.ReadFull(os.Stdin, body); err != nil {
			return err
		}
		var msg message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		switch msg.Kind {
		case "init":
			var err error
			switch mode {
			case "eof-after-init":
				err = send("response", msg.ID, map[string]bool{"ok": true})
			case "reject-init", "reject-init-eof":
				err = send("error", msg.ID, map[string]string{"message": "injected init failure"})
			}
			if err != nil {
				return err
			}
			if mode != "reject-init" {
				if err := os.Stdout.Close(); err != nil {
					return err
				}
			}
		case "exit-after-eof":
			// The parent sends this only after observing EOF, so no sleep
			// determines whether EOF precedes the natural process exit.
			var data struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				return err
			}
			os.Exit(data.Code)
		default:
			return fmt.Errorf("unexpected fixture message %q", msg.Kind)
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
