package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	var tests = []struct {
		input    string
		wantData Statement
		wantErr  bool
	}{
		{"SEND 2 XLM FROM master TO jennifer", &SendRequest{
			Amount:   "2",
			Currency: "XLM",
			From:     "master",
			To:       "jennifer",
		}, false},
		{"SEND 2 XLM FROM FROM TO jennifer", nil, true},
		{"SEND 2 FROM master TO jennifer", nil, true},
		{"SEND 2 XLM FROM master TO", nil, true},
		{"SEND 2 XLM FROM TO", nil, true},
		{"SEND XLM FROM TO", nil, true},
		{"SEND 2 XLM jennifer", nil, true},
		{"SEND 2 XLM TO jennifer", &SendRequest{
			Amount:   "2",
			Currency: "XLM",
			To:       "jennifer",
		}, false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			statement, err := Parse(test.input)
			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantData, statement)
			switch s := statement.(type) {
			case *SendRequest:
				require.Equal(t, SendKind, s.Kind())
			default:
				t.Fatalf("unexpected type %T", s)
			}
		})
	}
}
