package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

var (
	tokenEOF = &token{kind: tokenEof}
)

type lexer struct {
	reader *strings.Reader
}

func (l *lexer) Next() (*token, error) {
	if err := l.skipWhitespaces(); err != nil {
		if err == io.EOF {
			return tokenEOF, nil
		}

		return nil, err
	}

	r, _, err := l.reader.ReadRune()
	if err != nil {
		return nil, err
	}

	switch r {
	// check quotes, commas, etc..
	case ',':
		return &token{kind: tokenCOMMA, value: string(r)}, nil
	case '=':
		return &token{kind: tokenEQUAL, value: string(r)}, nil
	}

	if err := l.reader.UnreadRune(); err != nil {
		return nil, err
	}

	value, err := l.scanIdent()
	if err != nil {
		return nil, err
	}

	var kinds []tokenKind
	for i := _tokStartKeywords + 1; i < _tokEndKeywords; i++ {
		kinds = append(kinds, tokenKind(i))
	}

	for _, kind := range kinds {
		if strings.ToUpper(value) == kind.String() {
			return &token{kind: kind, value: value}, nil
		}
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return &token{kind: tokenNumber, value: value}, nil
	}

	if value == "" {
		return nil, fmt.Errorf("unable to find token for '%s'", value)
	}

	return &token{kind: tokenIdent, value: value}, nil
}

func (l *lexer) scanIdent() (string, error) {
	str := ""
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if !isNotAlphaNum(r) {
			l.reader.UnreadRune()
			break
		}

		str += string(r)
	}

	return str, nil
}

func isNotAlphaNum(r rune) bool {
	return !unicode.IsSpace(r) && !strings.ContainsRune("`=\"',", r)
}

func (l *lexer) skipWhitespaces() error {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			return err
		}

		if !unicode.IsSpace(r) {
			if err := l.reader.UnreadRune(); err != nil {
				return err
			}

			break
		}
	}

	return nil
}
