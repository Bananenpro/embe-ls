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

var keywords = []string{
	"if", "elif", "else", "while", "for", "var",
}

var types = []string{
	"number", "string", "boolean",
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

	return document.getCompletions(parts[len(parts)-1], int(pos.Line)), nil
}

func (d *Document) getCompletions(item string, line int) []protocol.CompletionItem {
	completions := make([]protocol.CompletionItem, 0)

	parts := strings.Split(item, ".")
	base := strings.Join(parts[:len(parts)-1], ".") + "."
	if len(parts) == 1 {
		base = ""
	}

	eventCompletionType := protocol.CompletionItemKindEvent
	for e := range generator.Events {
		if strings.HasPrefix("@"+e, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: e,
				Kind:  &eventCompletionType,
			})
		}
	}

	keywordCompletionType := protocol.CompletionItemKindKeyword
	for _, k := range keywords {
		if strings.HasPrefix(k, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(k, base),
				Kind:  &keywordCompletionType,
			})
		}
	}

	classCompletionType := protocol.CompletionItemKindClass
	for _, t := range types {
		if strings.HasPrefix(t, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(t, base),
				Kind:  &classCompletionType,
			})
		}
	}

	funcCompletionType := protocol.CompletionItemKindFunction
	for f := range generator.FuncCalls {
		if strings.HasPrefix(f, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(f, base),
				Kind:  &funcCompletionType,
			})
		}
	}

	varCompletionType := protocol.CompletionItemKindVariable
	for name, v := range d.variables {
		if v.Name.Line < line && strings.HasPrefix(name, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(name, base),
				Kind:  &varCompletionType,
			})
		}
	}

	constCompletionType := protocol.CompletionItemKindConstant
	for name, c := range d.constants {
		if c.Name.Line < line && strings.HasPrefix(name, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(name, base),
				Kind:  &constCompletionType,
			})
		}
	}

	for v := range generator.Variables {
		if strings.HasPrefix(v, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(v, base),
				Kind:  &varCompletionType,
			})
		}
	}

	for f := range generator.ExprFuncCalls {
		if strings.HasPrefix(f, item) {
			completions = append(completions, protocol.CompletionItem{
				Label: strings.TrimPrefix(f, base),
				Kind:  &funcCompletionType,
			})
		}
	}

	snippetCompletionType := protocol.CompletionItemKindSnippet
	snippetInsertTextFormat := protocol.InsertTextFormatSnippet
	for label, s := range snippets {
		if strings.HasPrefix(s, item) {
			snippet := s
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
