package parser

import (
	"fmt"
)

type tokenKind int

//go:generate stringer -type=tokenKind -linecomment
const (
	tokenUnknown tokenKind = iota
	tokenEof               // EOF

	tokenIdent  // IDENT
	tokenSTRING // STRING

	//

	_tokStartKeywords

	tokenSelect  // SELECT
	tokenSend    // SEND
	tokenSHARE   // SHARE
	tokenACCOUNT // ACCOUNT

	tokenFrom // FROM
	tokenTo   // TO

	tokenWith  // WITH
	tokenWhere // WHERE
	tokenAND   // AND
	tokenSET   // SET
	tokenDATA  // DATA

	_tokEndKeywords

	//

	tokenNumber // NUMBER
	tokenCOMMA  // COMMA
	tokenEQUAL  // EQUAL
	tokenQUOTES // QUOTES
)

type token struct {
	kind  tokenKind
	value string
}

func (t token) String() string {
	return fmt.Sprintf("%s('%s')", t.kind.String(), t.value)
}
