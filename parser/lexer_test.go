package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	var tests = []struct {
		input     string
		wantKind  tokenKind
		wantValue string
		wantErr   bool
	}{
		{"from", tokenFrom, "from", false},
		{"from ", tokenFrom, "from", false},
		{"to", tokenTo, "to", false},
		// {"`", 0, "", true},
		{"", tokenEof, "", false},
	}

	for _, test := range tests {
		l := lexer{reader: strings.NewReader(test.input)}
		tok, err := l.Next()
		if test.wantErr {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)

		require.Equal(t, test.wantKind, tok.kind, tok.kind.String())
		require.Equal(t, test.wantValue, tok.value, tok.value)
	}
}
