package wallet

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestManager(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	path := f.Name()
	require.NoError(t, f.Close())

	secret := "hello"

	m, err := Open(path, nil)
	require.NoError(t, err)

	kp, err := keypair.Random()
	require.NoError(t, err)

	w := New(kp.Address(), kp)
	err = m.AddWallet(w)
	require.NoError(t, err)

	err = Write(path, m)
	require.Error(t, err) // needs to be unlocked to write

	m.Unlock([]byte(secret))
	err = Write(path, m)
	require.NoError(t, err) // good to go

	//

	m, err = Open(path, nil)
	require.NoError(t, err)

	require.Equal(t, 1, len(m.Stellar.Wallets))
	gotKP, ok := m.Stellar.Wallets[0].Keypair.(*keypair.FromAddress)
	require.True(t, ok)
	require.Equal(t, kp.Address(), gotKP.Address())

	//

	m, err = Open(path, []byte(secret))
	require.NoError(t, err)
	gotFullKP, ok := m.Stellar.Wallets[0].Keypair.(*keypair.Full)
	require.True(t, ok)
	require.Equal(t, kp.Address(), gotFullKP.Address())
	require.Equal(t, kp.Seed(), gotFullKP.Seed())
}

func TestOpenssl(t *testing.T) {
	secret := []byte("secret")
	h := sha256.New()
	h.Write(secret)
	secret = h.Sum(nil)

	plaintext := []byte("plaintext")
	encrypted, err := encrypt(secret, plaintext)
	require.NoError(t, err)

	d, err := decrypt(secret, encrypted)
	require.NoError(t, err)
	require.Equal(t, string(plaintext), string(d))

	fmt.Println(base64.StdEncoding.EncodeToString(encrypted))
}

func TestMemo(t *testing.T) {
	want := Memo{
		Type: MEMO_HASH,
		Value: MemoValue{
			HashValue: xdr.Hash{1, 2, 3},
		},
	}
	b, err := yaml.Marshal(want)
	require.NoError(t, err)

	var got Memo
	err = yaml.Unmarshal(b, &got)
	require.NoError(t, err)

	require.Equal(t, want, got)
}

func BenchmarkRandom(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			keypair.Random()
		}
	})
}
