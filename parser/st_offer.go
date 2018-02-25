package parser

import (
	"fmt"
)

type Offer struct {
	Account string
	Amount  string
	Buying  string
	Selling string
	Price   string
	kind    Kind
}

func (s *Offer) Kind() Kind {
	return s.kind
}

func (s *Offer) parse(l *lexer) (err error) {
	s.Amount, err = parseExpect(l, tokenNumber)
	if err != nil {
		return err
	}

	cur, err := parseExpect(l, tokenIdent, tokenSTRING)
	if err != nil {
		return err
	}

	switch s.kind {
	case BuyOfferKind:
		s.Buying = cur
	case SellOfferKind:
		s.Selling = cur
	}

loop:
	for {
		tok, err := l.Next()
		if err != nil {
			return err
		}

		switch tok.kind {
		case tokenAT:
			s.Price, err = parseExpect(l, tokenNumber)
		case tokenUSING, tokenFOR:
			cur, err = parseExpect(l, tokenIdent, tokenSTRING)
			switch s.kind {
			case BuyOfferKind:
				s.Selling = cur
			case SellOfferKind:
				s.Buying = cur
			}
		case tokenWith:
			s.Account, err = parseExpect(l, tokenIdent, tokenSTRING)
		case tokenEof:
			break loop
		default:
			return fmt.Errorf("expected '%v' or '%v' but got '%s'", tokenAT, tokenUSING, tok)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func parseExpect(l *lexer, expected ...tokenKind) (string, error) {
	tok, err := l.Next()
	if err != nil {
		return "", err
	}

	switch {
	case containsKind(tok.kind, expected):
		return tok.value, nil
	default:
		return "", fmt.Errorf("expected '%v' but got '%s'", expected, tok)
	}
}

func containsKind(got tokenKind, list []tokenKind) bool {
	for _, want := range list {
		if want == got {
			return true
		}
	}

	return false
}
