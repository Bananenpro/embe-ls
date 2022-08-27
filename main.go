package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
)

var (
	name    = "embe-ls"
	version = "0.0.1"
)

var handler protocol.Handler

func main() {
	logging.Configure(1, nil)

	handler = protocol.Handler{
		Initialize:            initialize,
		Initialized:           initialized,
		Shutdown:              shutdown,
		SetTrace:              setTrace,
		TextDocumentDidOpen:   textDocumentDidOpen,
		TextDocumentDidChange: textDocumentDidChange,
		TextDocumentDidClose:  textDocumentDidClose,
	}

	var protocol string
	pflag.StringVarP(&protocol, "protocol", "p", "stdio", "The protocol to use. ('stdio', 'tcp', 'websocket', 'nodejs')")
	var address string
	pflag.StringVarP(&address, "address", "a", ":4389", "The address to use for a TCP or WebSocket protocol.")

	server := server.NewServer(&handler, name, false)

	var err error
	switch protocol {
	case "stdio":
		err = server.RunStdio()
	case "tcp":
		err = server.RunTCP(address)
	case "websocket":
		err = server.RunWebSocket(address)
	case "nodejs":
		err = server.RunNodeJs()
	default:
		err = fmt.Errorf("unsupported protocol: %s", protocol)
	}
	util.FailOnError(err)
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    name,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
