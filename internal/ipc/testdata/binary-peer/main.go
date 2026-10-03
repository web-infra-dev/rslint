// A real Go peer for the Node IPC integration suite. No lint or parser code is
// involved: one request carries text and arbitrary bytes across multiple slots.
package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"slices"

	"github.com/web-infra-dev/rslint/internal/ipc"
)

type payload struct {
	Round   int      `json:"round"`
	Hashes  []string `json:"hashes"`
	Lengths []int    `json:"lengths"`
}

func main() {
	channel := ipc.NewChannel(os.Stdin, os.Stdout)
	channel.SetInboundHandler(func(ctx context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind != "startBinary" {
			return nil, fmt.Errorf("unexpected fixture request %q", msg.Kind)
		}
		var request struct {
			Round int `json:"round"`
		}
		if err := msg.Decode(&request); err != nil {
			return nil, err
		}
		data := make([]byte, ipc.MemorySlotSize+257)
		for i := range data {
			data[i] = byte(i*131 + request.Round*17 + i>>8)
		}
		text := "\ufeff" + "arbitrary text: café 😀\r\n\x00"
		expected := payload{Round: request.Round}
		for _, bytes := range [][]byte{[]byte(text), {}, {}, data} {
			expected.Hashes = append(expected.Hashes, fmt.Sprintf("%x", sha256.Sum256(bytes)))
			expected.Lengths = append(expected.Lengths, len(bytes))
		}
		response, err := channel.SendRequest(ctx, "binaryAttachments", expected,
			ipc.Text(text), ipc.Text(""), ipc.Bytes([]byte{}), ipc.Bytes(data))
		if err != nil {
			return nil, err
		}
		var actual payload
		if err := response.Decode(&actual); err != nil {
			return nil, err
		}
		if actual.Round != expected.Round || !slices.Equal(actual.Hashes, expected.Hashes) || !slices.Equal(actual.Lengths, expected.Lengths) {
			return nil, fmt.Errorf("attachment data changed during round %d", request.Round)
		}
		return expected, nil
	})
	channel.Start()
	<-channel.Done()
	_ = channel.Close()
}
