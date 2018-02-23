package parser

type tokenKind int

//go:generate stringer -type=tokenKind -linecomment
const (
	tokenUnknown tokenKind = iota
	tokenEof

	tokenIdent // IDENT

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

	_tokEndKeywords

	//

	tokenNumber // NUMBER
	tokenCOMMA  // COMMA
)

type token struct {
	kind  tokenKind
	value string
}
