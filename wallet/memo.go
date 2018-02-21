package wallet

import (
	"fmt"
	"strconv"

	"github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
)

//go:generate stringer -type=MemoKind -linecomment
type MemoKind int

const (
	MEMO_TEXT   MemoKind = iota + 1 // text
	MEMO_ID                         // id
	MEMO_HASH                       // hash
	MEMO_RETURN                     // return
)

type Memo struct {
	Type  MemoKind  `yaml:"type,omitempty"`
	Value MemoValue `yaml:"value,omitempty"`
}

type MemoValue struct {
	HashValue   xdr.Hash
	IntValue    uint64
	StringValue string
}

func MemoFromString(kind MemoKind, str string) (*Memo, error) {
	memo := &Memo{
		Type: kind,
	}

	switch kind {
	case MEMO_TEXT:
		memo.Value.StringValue = str
	case MEMO_ID:
		id, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		memo.Value.IntValue = uint64(id)
	case MEMO_RETURN, MEMO_HASH:
		var hash xdr.Hash
		err := xdr.SafeUnmarshalBase64(str, &hash)
		if err != nil {
			return nil, err
		}
		memo.Value.HashValue = hash
	default:
		return nil, fmt.Errorf("unknown kind '%v'", kind)
	}

	return memo, nil
}

func (m Memo) ToTransactionMutator() build.TransactionMutator {
	switch m.Type {
	case MEMO_TEXT:
		return &build.MemoText{m.Value.StringValue}
	case MEMO_ID:
		return &build.MemoID{m.Value.IntValue}
	case MEMO_RETURN:
		return &build.MemoReturn{m.Value.HashValue}
	case MEMO_HASH:
		return &build.MemoHash{m.Value.HashValue}
	default:
		fmt.Printf("m.Type: %+v\n", m.Type)
		return nil
	}
}

func (v Memo) MarshalYAML() (interface{}, error) {
	o := map[string]interface{}{
		"type": v.Type.String(),
	}
	var i interface{}
	switch v.Type {
	case MEMO_TEXT:
		i = v.Value.StringValue
	case MEMO_ID:
		i = v.Value.IntValue
	case MEMO_RETURN, MEMO_HASH:
		str, err := xdr.MarshalBase64(v.Value.HashValue)
		if err != nil {
			return nil, err
		}
		i = str
	default:
		return nil, fmt.Errorf("unknown memo type: '%v'", v.Type)
	}

	o["value"] = i

	return o, nil
}

func (m *Memo) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var in map[string]interface{}
	if err := unmarshal(&in); err != nil {
		return err
	}

	v := in["value"]
	switch kind := in["type"]; kind {
	case MEMO_TEXT.String():
		m.Type = MEMO_TEXT
		m.Value.StringValue = v.(string)
	case MEMO_ID.String():
		m.Type = MEMO_ID
		m.Value.IntValue = uint64(v.(int))
	case MEMO_RETURN.String():
		m.Type = MEMO_RETURN
		var hash xdr.Hash
		err := xdr.SafeUnmarshalBase64(v.(string), &hash)
		if err != nil {
			return err
		}
		m.Value.HashValue = hash
	case MEMO_HASH.String():
		m.Type = MEMO_HASH
		var hash xdr.Hash
		err := xdr.SafeUnmarshalBase64(v.(string), &hash)
		if err != nil {
			return err
		}
		m.Value.HashValue = hash
	default:
		return fmt.Errorf("unknown memo type: '%v'", kind)
	}

	return nil
}
