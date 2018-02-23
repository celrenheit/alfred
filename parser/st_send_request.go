package parser

import "fmt"

type SendRequest struct {
	Amount   string
	Currency string
	From, To string
}

func (s *SendRequest) Kind() Kind {
	return SendKind
}

func (s *SendRequest) parse(l *lexer) error {
	for i := 0; i < 2; i++ {
		tok, err := l.Next()
		if err != nil {
			return err
		}

		switch tok.kind {
		case tokenNumber:
			s.Amount = tok.value
		case tokenIdent:
			s.Currency = tok.value
		default:
			return fmt.Errorf("unexpected token '%v' for '%s', should be AMOUNT CURRENCY", tok.kind, tok.value)
		}
	}

	var (
		tok *token
		err error
	)
	for tok, err = l.Next(); err == nil && tok.kind != tokenEof; tok, err = l.Next() {
		switch tok.kind {
		case tokenFrom:
			s.From, err = parseIdent(l)
		case tokenTo:
			s.To, err = parseIdent(l)
		default:
			return fmt.Errorf("unexpected token '%v' for '%s', should be only FROM and TO keywords", tok.kind, tok.value)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
