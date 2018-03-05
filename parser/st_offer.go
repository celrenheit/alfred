package parser

import (
	"fmt"
)

type Offer struct {
	Account    string
	Amount     string
	AmountKind AmountOfferKind
	Buying     string
	Selling    string
	Price      string
	kind       Kind
}

type AmountOfferKind int

const (
	AmountBuyKind AmountOfferKind = iota + 1
	AmountSellKind
)

func (s *Offer) Kind() Kind {
	return s.kind
}

func (s *Offer) parse(l *lexer) (err error) {
	tok, err := l.Next()
	if err != nil {
		return err
	}

	var cur string
	switch tok.kind {
	case tokenNumber:
		s.Amount = tok.value
		s.AmountKind = AmountBuyKind
	case tokenIdent, tokenSTRING:
		cur = tok.value
	default:
		return fmt.Errorf("expected amount or currency but got '%s'", tok)
	}

	if cur == "" {
		cur, err = parseExpect(l, tokenIdent, tokenSTRING)
		if err != nil {
			return err
		}
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
		checkCurrency:
			tok, err := parseTokenExpect(l, tokenIdent, tokenSTRING, tokenNumber)
			if err != nil {
				return err
			}
			switch tok.kind {
			case tokenSTRING, tokenIdent:
				cur := tok.value
				switch s.kind {
				case BuyOfferKind:
					s.Selling = cur
				case SellOfferKind:
					s.Buying = cur
				}
			case tokenNumber:
				if s.Amount != "" {
					return fmt.Errorf("only one amount is allowed (previous : %v)", s.Amount)
				}

				s.Amount = tok.value
				s.AmountKind = AmountSellKind
				goto checkCurrency
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
	tok, err := parseTokenExpect(l, expected...)
	if err != nil {
		return "", err
	}

	return tok.value, nil
}

func parseTokenExpect(l *lexer, expected ...tokenKind) (*token, error) {
	tok, err := l.Next()
	if err != nil {
		return nil, err
	}

	switch {
	case containsKind(tok.kind, expected):
		return tok, nil
	default:
		return nil, fmt.Errorf("expected '%v' but got '%s'", expected, tok)
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
