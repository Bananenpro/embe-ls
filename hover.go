package main

import (
	"fmt"

	"github.com/Bananenpro/embe/generator"
	"github.com/Bananenpro/embe/parser"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentHover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	document, ok := getDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	var token parser.Token
	var tokenIndex int
	for i, t := range document.tokens {
		if t.Line == int(params.Position.Line) && int(params.Position.Character) >= t.Column && int(params.Position.Character) <= t.Column+len(t.Lexeme) {
			token = t
			tokenIndex = i
			break
		}
	}
	if token.Type != parser.TkIdentifier {
		return nil, nil
	}

	identifierName := token.Lexeme

	var signature string

	for _, f := range document.functions {
		if int(params.Position.Line) >= f.StartLine && int(params.Position.Line) <= f.EndLine {
			for _, p := range f.Params {
				signature = fmt.Sprintf("var %s: %s", p.Name.Lexeme, p.Type.DataType)
			}
		}
	}

	if signature == "" {
		if e, ok := generator.Events[token.Lexeme]; ok && tokenIndex > 0 && document.tokens[tokenIndex-1].Type == parser.TkAt {
			signature = e.String()
			identifierName = "@" + identifierName
		} else if f, ok := generator.FuncCalls[token.Lexeme]; ok {
			paramCount := getParamCount(document.tokens, tokenIndex+2)
			for _, s := range f.Signatures {
				if len(s.Params) == paramCount {
					signature = "func " + s.String()
					break
				}
			}
		} else if ef, ok := generator.ExprFuncCalls[token.Lexeme]; ok {
			paramCount := getParamCount(document.tokens, tokenIndex+2)
			for _, s := range ef.Signatures {
				if len(s.Params) == paramCount {
					signature = "func " + s.String()
					break
				}
			}
		} else if v, ok := generator.Variables[token.Lexeme]; ok {
			signature = v.String()
		} else if cv, ok := document.variables[token.Lexeme]; ok {
			signature = fmt.Sprintf("var %s: %s", cv.Name.Lexeme, cv.DataType)
		} else if c, ok := document.constants[token.Lexeme]; ok {
			signature = fmt.Sprintf("const %s: %s = %s", c.Name.Lexeme, c.Type, c.Value.Lexeme)
		} else if cf, ok := document.functions[token.Lexeme]; ok {
			signature = "func " + cf.Name.Lexeme + "("
			for i, p := range cf.Params {
				if i > 0 {
					signature += ", "
				}
				signature += p.Name.Lexeme + ": " + string(p.Type.DataType)
			}
			signature += ")"
		}
	}

	if signature == "" {
		return nil, nil
	}

	value := fmt.Sprintf("```embe\n%s\n```", signature)

	if docs, ok := documentation[identifierName]; ok {
		value += "\n---\n" + docs
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: value,
		},
	}, nil
}

func getParamCount(tokens []parser.Token, start int) int {
	parens := 1
	paramCount := 1
	for i := start; i < len(tokens) && parens > 0 && tokens[i].Type != parser.TkNewLine; i++ {
		switch tokens[i].Type {
		case parser.TkOpenParen:
			parens++
		case parser.TkCloseParen:
			parens--
		case parser.TkComma:
			if parens == 1 {
				paramCount++
			}
		}
	}
	if parens != 0 || (start < len(tokens) && tokens[start].Type == parser.TkCloseParen) {
		return 0
	}
	return paramCount
}