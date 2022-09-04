package main

import (
	"bytes"
	"sync"

	"github.com/Bananenpro/embe/blocks"
	"github.com/Bananenpro/embe/generator"
	"github.com/Bananenpro/embe/parser"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/kutil/logging"
)

type Document struct {
	uri         protocol.DocumentUri
	content     string
	changed     bool
	diagnostics []protocol.Diagnostic
	variables   map[string]*blocks.Variable
	constants   map[string]*generator.Constant
}

var documents sync.Map

func (d *Document) validate(notify glsp.NotifyFunc) {
	if !d.changed {
		return
	}
	d.changed = false

	defer d.sendDiagnostics(notify)

	severityWarning := protocol.DiagnosticSeverityWarning
	severityError := protocol.DiagnosticSeverityError

	d.diagnostics = d.diagnostics[:0]

	tokens, lines, err := parser.Scan(bytes.NewBufferString(d.content))
	if err != nil {
		if e, ok := err.(parser.ScanError); ok {
			d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      uint32(e.Line),
						Character: uint32(e.Column),
					},
					End: protocol.Position{
						Line:      uint32(e.Line),
						Character: uint32(e.Column + 1),
					},
				},
				Severity: &severityError,
				Message:  e.Message,
			})
		} else {
			logging.GetLogger(name).Errorf("Failed to scan '%s': %s", d.uri, err)
		}
		return
	}

	statements, errs := parser.Parse(tokens, lines)
	if len(errs) > 0 {
		for _, err := range errs {
			if e, ok := err.(parser.ParseError); ok {
				d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(e.Token.Line),
							Character: uint32(e.Token.Column),
						},
						End: protocol.Position{
							Line:      uint32(e.Token.Line),
							Character: uint32(e.Token.Column + len(e.Token.Lexeme)),
						},
					},
					Severity: &severityError,
					Message:  e.Message,
				})
			} else {
				logging.GetLogger(name).Errorf("Failed to parse '%s': %s", d.uri, err)
			}
		}
		return
	}

	_, variables, constants, warnings, errs := generator.GenerateBlocks(statements, lines)
	for _, warning := range warnings {
		if w, ok := warning.(generator.GenerateError); ok {
			d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      uint32(w.Token.Line),
						Character: uint32(w.Token.Column),
					},
					End: protocol.Position{
						Line:      uint32(w.Token.Line),
						Character: uint32(w.Token.Column + len(w.Token.Lexeme)),
					},
				},
				Severity: &severityWarning,
				Message:  w.Message,
			})
		} else {
			logging.GetLogger(name).Errorf("Failed to generate blocks for '%s': %s", d.uri, err)
		}
	}
	if len(errs) > 0 {
		for _, err := range errs {
			if e, ok := err.(generator.GenerateError); ok {
				d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(e.Token.Line),
							Character: uint32(e.Token.Column),
						},
						End: protocol.Position{
							Line:      uint32(e.Token.Line),
							Character: uint32(e.Token.Column + len(e.Token.Lexeme)),
						},
					},
					Severity: &severityError,
					Message:  e.Message,
				})
			} else {
				logging.GetLogger(name).Errorf("Failed to generate blocks for '%s': %s", d.uri, err)
			}
		}
		return
	}
	d.variables = variables
	d.constants = constants
}

func (d *Document) sendDiagnostics(notify glsp.NotifyFunc) {
	notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         d.uri,
		Diagnostics: d.diagnostics,
	})
}

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	document := &Document{
		uri:         params.TextDocument.URI,
		content:     params.TextDocument.Text,
		changed:     true,
		diagnostics: make([]protocol.Diagnostic, 0),
		variables:   make(map[string]*blocks.Variable),
	}
	documents.Store(params.TextDocument.URI, document)
	go document.validate(context.Notify)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if document, ok := getDocument(params.TextDocument.URI); ok {
		content := document.content
		for _, change := range params.ContentChanges {
			if c, ok := change.(protocol.TextDocumentContentChangeEvent); ok {
				start, end := c.Range.IndexesIn(content)
				content = content[:start] + c.Text + content[end:]
			} else if c, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
				content = c.Text
			}
		}
		document.content = content
		document.changed = len(params.ContentChanges) > 0
		go document.validate(context.Notify)
	}
	return nil
}

func textDocumentDidClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	_, ok := documents.LoadAndDelete(params.TextDocument.URI)
	if ok {
		go context.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: make([]protocol.Diagnostic, 0),
		})
	}
	return nil
}

func getDocument(uri protocol.DocumentUri) (*Document, bool) {
	doc, ok := documents.Load(uri)
	if !ok {
		return nil, false
	}
	return doc.(*Document), true
}
