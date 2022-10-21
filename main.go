package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"
	_ "github.com/tliron/kutil/logging/simple"

	"github.com/Bananenpro/embe-ls/config"
	"github.com/Bananenpro/embe-ls/log"
)

var (
	name    = "embe-ls"
	version = "0.1.3"
)

var handler protocol.Handler

func main() {
	log.Info("Starting %s v%s...", name, version)
	glspLogLevel := 0
	if config.GLSPLogFile != nil {
		glspLogLevel = 2
	}
	logging.Configure(glspLogLevel, config.GLSPLogFile)

	handler = protocol.Handler{
		Initialize:                    initialize,
		Initialized:                   initialized,
		Shutdown:                      shutdown,
		SetTrace:                      setTrace,
		TextDocumentDidOpen:           textDocumentDidOpen,
		TextDocumentDidChange:         textDocumentDidChange,
		TextDocumentDidClose:          textDocumentDidClose,
		TextDocumentCompletion:        textDocumentCompletion,
		TextDocumentSignatureHelp:     textDocumentSignatureHelp,
		TextDocumentHover:             textDocumentHover,
		TextDocumentColor:             textDocumentColor,
		TextDocumentColorPresentation: textDocumentColorPresentation,
	}

	var protocol string
	pflag.StringVarP(&protocol, "protocol", "p", "stdio", "The protocol to use. ('stdio', 'tcp', 'websocket', 'node-ipc')")
	var address string
	pflag.StringVarP(&address, "address", "a", ":4389", "The address to use for a TCP or WebSocket protocol.")
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Parse()

	server := server.NewServer(&handler, name, config.GLSPLogFile != nil)

	var err error
	switch protocol {
	case "stdio":
		log.Info("Protocol: STDIO")
		err = server.RunStdio()
	case "tcp":
		log.Info("Protocol: TCP")
		err = server.RunTCP(address)
	case "websocket":
		log.Info("Protocol: WebSocket")
		err = server.RunWebSocket(address)
	case "node-ipc":
		log.Info("Protocol: Node IPC")
		err = server.RunNodeJs()
	default:
		err = fmt.Errorf("Unsupported protocol: %s", protocol)
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	log.Trace("Initializing capabilities...")
	capabilities := handler.CreateServerCapabilities()
	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindIncremental
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{"@", "#", "."},
	}
	capabilities.SignatureHelpProvider = &protocol.SignatureHelpOptions{
		TriggerCharacters: []string{"(", ","},
	}
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    name,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	log.Trace("Initialized.")
	return nil
}

func shutdown(context *glsp.Context) error {
	log.Info("Shutdown.")
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
