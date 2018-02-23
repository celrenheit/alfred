package parser

import (
	"errors"
)

type Kind int

const (
	SendKind Kind = iota + 1
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
		return "", errors.New("should be an identifier token")
	}

	return tok.value, nil
}
