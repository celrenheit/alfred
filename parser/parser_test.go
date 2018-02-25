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
		{"SHARE ACCOUNT master WITH alice, bob, celine", &ShareAccountRequest{
			Account:            "master",
			AdditionnalSigners: []string{"alice", "bob", "celine"},
		}, false},
		{"SHARE ACCOUNT master WITH alice, bob and celine", &ShareAccountRequest{
			Account:            "master",
			AdditionnalSigners: []string{"alice", "bob", "celine"},
		}, false},
		{"SHARE ACCOUNT master WITH alice and bob, celine", &ShareAccountRequest{
			Account:            "master",
			AdditionnalSigners: []string{"alice", "bob", "celine"},
		}, false},
		{"SHARE ACCOUNT WITH alice, bob, celine", nil, true},
		{"SHARE ACCOUNT master WITH alice,,, bob, celine", nil, true},
		{"SHARE ACCOUNT master WITH alice,,", nil, true},
		{"SHARE ACCOUNT master WITH", nil, true},
		{"SHARE ACCOUNT master WITH ,", nil, true},
		{"SHARE ACCOUNT master", nil, true},
		{"SET DATA foo = bar", &SetDataRequest{
			KVs: map[string]DataEntry{
				"foo": {SetDataFromString, "bar"},
			},
		}, false},
		{`SET DATA foo = "hello world"`, &SetDataRequest{
			KVs: map[string]DataEntry{
				"foo": {SetDataFromString, "hello world"},
			},
		}, false},
		{"SET DATA foo from ./bar", &SetDataRequest{
			KVs: map[string]DataEntry{
				"foo": {SetDataFromFile, "./bar"},
			},
		}, false},
		{`SET DATA "hello world" from "./bar"`, &SetDataRequest{
			KVs: map[string]DataEntry{
				"hello world": {SetDataFromFile, "./bar"},
			},
		}, false},
		{`SET DATA foo from ./text.txt, bar = "hello world"`, &SetDataRequest{
			KVs: map[string]DataEntry{
				"foo": {SetDataFromFile, "./text.txt"},
				"bar": {SetDataFromString, "hello world"},
			},
		}, false},
		{"SET DATA foo", nil, true},
		{"SET DATA foo = ", nil, true},
		{`SET DATA foo = "`, nil, true},
		{`SET DATA foo from`, nil, true},
		{`BUY 100 MOBI AT 0.1000 USING XLM`, &Offer{
			kind:    BuyOfferKind,
			Amount:  "100",
			Buying:  "MOBI",
			Price:   "0.1000",
			Selling: "XLM",
		}, false},
		{`BUY 100 MOBI`, &Offer{ // defaults to XLM at the best price
			kind:    BuyOfferKind,
			Amount:  "100",
			Buying:  "MOBI",
			Price:   "",
			Selling: "",
		}, false},
		{`SELL 100 MOBI AT 0.1000 FOR XLM with wallet1`, &Offer{
			kind:    SellOfferKind,
			Amount:  "100",
			Buying:  "XLM",
			Price:   "0.1000",
			Selling: "MOBI",
			Account: "wallet1",
		}, false},
		{`BUY 100 MOBI USING AT`, nil, true},
		{`BUY MOBI 100`, nil, true},
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
			// switch s := statement.(type) {
			// case *SendRequest:
			// 	require.Equal(t, SendKind, s.Kind())
			// case *ShareAccountRequest:
			// 	require.Equal(t, ShareAccountKind, s.Kind())
			// case *SetDataRequest:
			// 	require.Equal(t, SetDataKind, s.Kind())
			// default:
			// 	t.Fatalf("unexpected type %T", s)
			// }
		})
	}
}
