package parser

import (
	"fmt"
)

type ShareAccountRequest struct {
	Account            string
	AdditionnalSigners []string
}

func (s *ShareAccountRequest) Kind() Kind {
	return ShareAccountKind
}

func (s *ShareAccountRequest) parse(l *lexer) error {
	tok, err := l.Next()
	if err != nil {
		return err
	}

	switch tok.kind {
	case tokenACCOUNT:
	default:
		return fmt.Errorf("unexpected token '%v' for '%s', should be ACCOUNT", tok.kind, tok.value)
	}

	s.Account, err = parseIdent(l)
	if err != nil {
		return err
	}

	tok, err = l.Next()
	if err != nil {
		return err
	}

	switch tok.kind {
	case tokenWith:
	default:
		return fmt.Errorf("unexpected token '%v' for '%s', should be WITH", tok.kind, tok.value)
	}

	s.AdditionnalSigners, err = parseList(l, tokenIdent)

	return err
}

func parseList(l *lexer, kind tokenKind) (list []string, err error) {

	var (
		tok *token
	)

loop:
	for {
		// account

		tok, err = l.Next()
		if err != nil {
			return nil, err
		}

		switch tok.kind {
		case kind:
			list = append(list, tok.value)
		case tokenEof:
			if len(list) > 0 {
				return nil, fmt.Errorf("unexpected token: '%v', should be: '%v'", tok, tokenIdent)
			}
			break loop
		default:
			return nil, fmt.Errorf("unexpected token: '%v', should be: '%v'", tok, tokenIdent)
		}

		// separator

		tok, err = l.Next()
		if err != nil {
			return nil, err
		}

		switch tok.kind {
		case tokenCOMMA, tokenAND:
			// continue
		case tokenEof:
			break loop
		default:
			return nil, fmt.Errorf("unexpected token: '%v', should be: '%v' or '%v'", tok, tokenCOMMA, tokenAND)
		}
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("got empty list")
	}

	return
}
