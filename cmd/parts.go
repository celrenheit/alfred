package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/stellar/go/build"
	"github.com/stellar/go/keypair"
)

var (
	limit        = 64
	headerPrefix = "alfred:"
)

func sha256sum(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

type KVData interface {
	Key() string
	Value() []byte
}

type Header struct {
	Filename string
	Length   uint64
	Checksum []byte
}

func (h *Header) Key() string {
	return headerPrefix + h.Filename
}

func (h *Header) Value() []byte {
	b := make([]byte, 8+32)
	binary.BigEndian.PutUint64(b[:8], h.Length)

	copy(b[8:], h.Checksum)

	return b
}

func ParseHeader(key string, value []byte) (*Header, error) {
	name := strings.TrimPrefix(key, headerPrefix)

	h := &Header{
		Filename: name,
		Length:   binary.BigEndian.Uint64(value[:8]),
		Checksum: make([]byte, 32),
	}

	copy(h.Checksum, value[8:])

	return h, nil
}

func ParsePart(prefix string, key string, value []byte) (*Part, error) {
	if !strings.HasPrefix(key, prefix) {
		return nil, fmt.Errorf("key: '%s' does not have prefix: '%s'", key, prefix)
	}

	offsetStr := strings.TrimPrefix(key, prefix+":")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return nil, err
	}

	return &Part{
		Filename: prefix,
		Offset:   uint32(offset),
		Data:     value,
	}, nil
}

func SplitIntoParts(filename string, data []byte) []*Part {
	parts := make([]*Part, len(data)/limit+1)
	var i uint32
	for len(data) >= limit {
		parts[i] = &Part{
			Filename: filename,
			Offset:   i,
			Data:     data[:limit],
		}
		i++
		data = data[limit:]
	}

	if len(data) > 0 {
		parts[i] = &Part{
			Filename: filename,
			Offset:   i,
			Data:     data,
		}
	}

	return parts
}

type Part struct {
	Filename string
	Offset   uint32
	Data     []byte
}

func (p *Part) Key() string {
	return p.Filename + ":" + strconv.Itoa(int(p.Offset))
}

func (p *Part) Value() []byte {
	return p.Data
}

func submitData(testnet, yes bool, src *keypair.Full, kvs []KVData) error {
	var sopts []build.TransactionMutator
	for _, kv := range kvs {
		sopts = append(sopts, build.SetData(kv.Key(), kv.Value()))
	}

	client := getClient(testnet)

	opts := []build.TransactionMutator{
		build.SourceAccount{src.Seed()},
		build.AutoSequence{SequenceProvider: client},
	}

	opts = append(opts, sopts...)

	if testnet {
		opts = append(opts, build.TestNetwork)
	} else {
		opts = append(opts, build.PublicNetwork)
	}

	tx, err := build.Transaction(opts...)
	if err != nil {
		return err
	}

	txe, err := tx.Sign(src.Seed())
	if err != nil {
		return err
	}

	txeB64, err := txe.Base64()
	if err != nil {
		return err
	}

	if !yes {
		_, err = (&promptui.Prompt{
			Label:     "Are you sure",
			IsConfirm: true,
		}).Run()
		if err != nil {
			return err
		}
	}

	resp, err := client.SubmitTransaction(txeB64)
	if err != nil {
		return err
	}

	fmt.Println(resp.Hash)
	return nil
}

func GetData(header *Header, parts []*Part) ([]byte, error) {
	data := make([]byte, header.Length)
	for _, p := range parts {
		offset := int(p.Offset) * 64
		copy(data[offset:offset+len(p.Data)], p.Data)
	}

	if !bytes.Equal(header.Checksum, sha256sum(data)) {
		return nil, fmt.Errorf("checksum invalid")
	}

	return data, nil
}
