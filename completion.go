package main

import (
	"strings"

	"github.com/Bananenpro/embe/generator"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	document, ok := getDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	pos := params.Position
	pos.Character = 0
	lineIndex := pos.IndexIn(document.content)
	endOfLineIndex := pos.EndOfLineIn(document.content)

	line := strings.TrimSpace(document.content[lineIndex:endOfLineIndex.IndexIn(document.content)])
	parts := strings.Split(line, " ")

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

	return completions
}
