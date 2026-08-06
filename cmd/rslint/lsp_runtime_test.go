package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

func TestConnectEditorRuntimeAuthenticatesLoopback(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	token := strings.Repeat("ab", 32)
	t.Setenv("RSLINT_RUNTIME_IPC_ADDRESS", listener.Addr().String())
	t.Setenv("RSLINT_RUNTIME_IPC_TOKEN", token)
	authenticated := make(chan string, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			authenticated <- "accept: " + acceptErr.Error()
			return
		}
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		line, readErr := bufio.NewReader(conn).ReadString('\n')
		if readErr != nil {
			authenticated <- "read: " + readErr.Error()
			return
		}
		authenticated <- line
	}()

	conn, err := connectEditorRuntime()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if got := <-authenticated; got != token+"\n" {
		t.Fatalf("authentication = %q, want token line", got)
	}
}

func TestConnectEditorRuntimeRejectsUnsafeEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		address string
		token   string
		want    string
	}{
		{name: "hostname", address: "localhost:1234", token: strings.Repeat("a", 64), want: "numeric loopback"},
		{name: "non-loopback", address: "192.0.2.1:1234", token: strings.Repeat("a", 64), want: "numeric loopback"},
		{name: "malformed-token", address: "127.0.0.1:1234", token: "not-a-token", want: "token"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("RSLINT_RUNTIME_IPC_ADDRESS", test.address)
			t.Setenv("RSLINT_RUNTIME_IPC_TOKEN", test.token)
			if conn, err := connectEditorRuntime(); err == nil {
				conn.Close()
				t.Fatal("unsafe endpoint was accepted")
			} else if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %q, want %q", err, test.want)
			}
		})
	}
}
