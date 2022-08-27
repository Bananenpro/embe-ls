package main

import (
	"regexp"
	"strings"

	"github.com/Bananenpro/embe/generator"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var snippets = map[string]string{
	"if statement":    "if ${1:condition}:\n  $0",
	"while loop":      "while ${1:condition}:\n  $0",
	"for loop":        "for ${1:count}:\n  $0",
	"var declaration": "var ${1:name}: ${2:type} = ${3:value}",
}

var completionSplitRegex = regexp.MustCompile(`[ (<>,!|&+\-\*/%=]`)

func textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	document, ok := getDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	pos := params.Position
	pos.Character = 0
	lineIndex := pos.IndexIn(document.content)

	line := strings.TrimSpace(document.content[lineIndex:params.Position.IndexIn(document.content)])
	parts := completionSplitRegex.Split(line, -1)

	return document.getCompletions(parts[len(parts)-1]), nil
}

func (d *Document) getCompletions(item string) []protocol.CompletionItem {
	completions := make([]protocol.CompletionItem, 0)

	eventCompletionType := protocol.CompletionItemKindEvent
	for e := range generator.Events {
		if strings.HasPrefix("@"+e, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: e,
				Kind:  &eventCompletionType,
			})
		}
	}

	funcCompletionType := protocol.CompletionItemKindFunction
	for f := range generator.FuncCalls {
		if strings.HasPrefix(f, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: f,
				Kind:  &funcCompletionType,
			})
		}
	}

	varCompletionType := protocol.CompletionItemKindVariable
	for v := range generator.Variables {
		if strings.HasPrefix(v, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: v,
				Kind:  &varCompletionType,
			})
		}
	}

	for f := range generator.ExprFuncCalls {
		if strings.HasPrefix(f, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: f,
				Kind:  &funcCompletionType,
			})
		}
	}

	snippetCompletionType := protocol.CompletionItemKindSnippet
	snippetInsertTextFormat := protocol.InsertTextFormatSnippet
	for label, s := range snippets {
		snippet := s
		if strings.HasPrefix(s, item) {
			completions = append(completions, protocol.CompletionItem{
				Label:            label,
				InsertText:       &snippet,
				InsertTextFormat: &snippetInsertTextFormat,
				Kind:             &snippetCompletionType,
			})
		}
	}

	return completions
}
