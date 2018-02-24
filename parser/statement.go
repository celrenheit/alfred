package parser

import (
	"fmt"
)

type Kind int

const (
	SendKind Kind = iota + 1
	ShareAccountKind
	SetDataKind
)

type Statement interface {
	Kind() Kind
	parse(l *lexer) error
}

func parseIdent(l *lexer) (string, error) {
	tok, err := l.Next()
	if err != nil {
		return "", err
	}

	if tok.kind != tokenIdent {
		return "", fmt.Errorf("should be an identifier token, got: '%v'", tok)
	}

	return tok.value, nil
}
