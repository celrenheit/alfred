package parser

import (
	"fmt"
	"strings"
)

func Parse(in string) (Statement, error) {
	return ParseReader(strings.NewReader(in))
}

func ParseReader(reader *strings.Reader) (s Statement, err error) {
	l := &lexer{reader}

	tok, err := l.Next()
	if err != nil {
		return nil, err
	}

	switch tok.kind {
	case tokenSend:
		s = &SendRequest{}
	case tokenSHARE:
		s = &ShareAccountRequest{}
	default:
		return nil, fmt.Errorf("parser: unknown statement '%s' got: '%v'", tok.value, tok)
	}

	err = s.parse(l)
	return s, err
}
