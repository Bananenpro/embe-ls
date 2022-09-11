package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Bananenpro/embe/generator"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var snippets = map[string]string{
	"if statement":      "if ${1:condition}:\n  $0",
	"while loop":        "while ${1:condition}:\n  $0",
	"for loop":          "for ${1:count}:\n  $0",
	"var declaration":   "var ${1:name}: ${2:type} = ${3:value}",
	"const declaration": "const ${1:name}: ${2:type} = ${3:value}",
	"func declaration":  "func ${1:name}(${2:parameters}):\n  $0",
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
	for _, e := range generator.Events {
		if strings.HasPrefix("@"+e.Name, item) {
			detail := e.String()
			completions = append(completions, protocol.CompletionItem{
				Label:         e.Name,
				Kind:          &eventCompletionType,
				Detail:        &detail,
				Documentation: getDocs("@" + e.Name),
			})
		}
	}

	keywordCompletionType := protocol.CompletionItemKindKeyword
	for _, k := range keywords {
		if strings.HasPrefix(k, item) {
			detail := k
			completions = append(completions, protocol.CompletionItem{
				Label:         strings.TrimPrefix(k, base),
				Kind:          &keywordCompletionType,
				Detail:        &detail,
				Documentation: getDocs(k),
			})
		}
	}

	classCompletionType := protocol.CompletionItemKindClass
	for _, t := range types {
		if strings.HasPrefix(t, item) {
			detail := t
			completions = append(completions, protocol.CompletionItem{
				Label:         strings.TrimPrefix(t, base),
				Kind:          &classCompletionType,
				Detail:        &detail,
				Documentation: getDocs(t),
			})
		}
	}

	funcCompletionType := protocol.CompletionItemKindFunction
	for _, f := range generator.FuncCalls {
		if strings.HasPrefix(f.Name, item) {
			detail := "func " + f.Signatures[0].String()
			completions = append(completions, protocol.CompletionItem{
				Label:         strings.TrimPrefix(f.Name, base),
				Kind:          &funcCompletionType,
				Detail:        &detail,
				Documentation: getDocs(f.Name),
			})
		}
	}

	varCompletionType := protocol.CompletionItemKindVariable
	parameters := make(map[string]struct{})
	for _, f := range d.functions {
		if line >= f.StartLine && line <= f.EndLine {
			for _, p := range f.Params {
				fmt.Fprintf(os.Stderr, "[%s]", p.Name.Lexeme)
				detail := fmt.Sprintf("var %s: %s", p.Name.Lexeme, p.Type.DataType)
				parameters[p.Name.Lexeme] = struct{}{}
				completions = append(completions, protocol.CompletionItem{
					Label:  strings.TrimPrefix(p.Name.Lexeme, base),
					Kind:   &varCompletionType,
					Detail: &detail,
				})
			}
		}
	}

	for _, f := range d.functions {
		if f.Name.Line < line && strings.HasPrefix(f.Name.Lexeme, item) {
			if _, ok := parameters[f.Name.Lexeme]; ok {
				continue
			}
			detail := "func " + f.Name.Lexeme + "("
			for i, p := range f.Params {
				if i > 0 {
					detail += ", "
				}
				detail += p.Name.Lexeme + ": " + string(p.Type.DataType)
			}
			detail += ")"
			completions = append(completions, protocol.CompletionItem{
				Label:  strings.TrimPrefix(f.Name.Lexeme, base),
				Kind:   &funcCompletionType,
				Detail: &detail,
			})
		}
	}

	for _, v := range d.variables {
		if v.Name.Line < line && strings.HasPrefix(v.Name.Lexeme, item) {
			if _, ok := parameters[v.Name.Lexeme]; ok {
				continue
			}
			detail := fmt.Sprintf("var %s: %s", v.Name.Lexeme, v.DataType)
			completions = append(completions, protocol.CompletionItem{
				Label:  strings.TrimPrefix(v.Name.Lexeme, base),
				Kind:   &varCompletionType,
				Detail: &detail,
			})
		}
	}

	constCompletionType := protocol.CompletionItemKindConstant
	for _, c := range d.constants {
		if c.Name.Line < line && strings.HasPrefix(c.Name.Lexeme, item) {
			if _, ok := parameters[c.Name.Lexeme]; ok {
				continue
			}
			detail := fmt.Sprintf("const %s: %s = %s", c.Name.Lexeme, c.Value.DataType, c.Value.Lexeme)
			completions = append(completions, protocol.CompletionItem{
				Label:  strings.TrimPrefix(c.Name.Lexeme, base),
				Kind:   &constCompletionType,
				Detail: &detail,
			})
		}
	}

	for _, v := range generator.Variables {
		if strings.HasPrefix(v.Name, item) {
			detail := v.String()
			completions = append(completions, protocol.CompletionItem{
				Label:         strings.TrimPrefix(v.Name, base),
				Kind:          &varCompletionType,
				Detail:        &detail,
				Documentation: getDocs(v.Name),
			})
		}
	}

	for _, f := range generator.ExprFuncCalls {
		if strings.HasPrefix(f.Name, item) {
			detail := "func " + f.Signatures[0].String()
			completions = append(completions, protocol.CompletionItem{
				Label:         strings.TrimPrefix(f.Name, base),
				Kind:          &funcCompletionType,
				Detail:        &detail,
				Documentation: getDocs(f.Name),
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
				Detail:           &snippet,
			})
		}
	}

	return completions
}
