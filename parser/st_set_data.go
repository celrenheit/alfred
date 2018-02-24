package parser

import (
	"fmt"
)

type SetDataEntryKind string

const (
	SetDataFromString SetDataEntryKind = "from-string"
	SetDataFromFile                    = "from-file"
)

type SetDataRequest struct {
	Account string
	KVs     map[string]DataEntry
}

type DataEntry struct {
	Kind  SetDataEntryKind
	Value string
}

func (s *SetDataRequest) Kind() Kind {
	return SetDataKind
}

func (s *SetDataRequest) parse(l *lexer) (err error) {

	s.KVs, err = parseOperations(l)

	return err
}

func parseOperations(l *lexer) (entries map[string]DataEntry, err error) {
	entries = map[string]DataEntry{}
	var (
		tok *token
	)

loop:
	for {
		tok, err = l.Next()
		if err != nil {
			return nil, err
		}

		var key string
		switch tok.kind {
		case tokenIdent, tokenSTRING:
			key = tok.value
		case tokenEof:
			if len(entries) > 0 {
				return nil, fmt.Errorf("unexpected token: '%v', should be: '%v'", tok, tokenIdent)
			}
			break loop
		default:
			return nil, fmt.Errorf("unexpected token: '%v', should be: '%v'", tok, tokenIdent)
		}

		// operation

		tok, err = l.Next()
		if err != nil {
			return nil, err
		}

		var op SetDataEntryKind
		switch tok.kind {
		case tokenEQUAL:
			op = SetDataFromString
		case tokenFrom:
			op = SetDataFromFile
		default:
			return nil, fmt.Errorf("unexpected token: '%v', should be: '%v'", tok, tokenIdent)
		}

		// value

		tok, err = l.Next()
		if err != nil {
			return nil, err
		}

		var value string
		switch tok.kind {
		case tokenIdent, tokenSTRING:
			value = tok.value
		default:
			return nil, fmt.Errorf("unexpected token: '%v', should be: '%v'", tok, tokenIdent)
		}

		if _, ok := entries[key]; ok {
			return nil, fmt.Errorf("'%v' set twice", key)
		}
		entries[key] = DataEntry{
			Kind:  op,
			Value: value,
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

	if len(entries) == 0 {
		return nil, fmt.Errorf("got empty list")
	}

	return
}
