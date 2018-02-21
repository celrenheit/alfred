package parser

type tokenKind int

//go:generate stringer -type=tokenKind -linecomment
const (
	tokenUnknown tokenKind = iota
	tokenEof

	tokenIdent // IDENT

	tokenSelect // SELECT
	tokenSend   // SEND

	tokenFrom // FROM
	tokenTo   // TO

	tokenWith  // WITH
	tokenWhere // WHERE

	tokenNumber // NUMBER
	tokenCOMMA  // COMMA
)

type token struct {
	kind  tokenKind
	value string
}
