package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/andreburgaud/crypt2go/padding"
	"github.com/stellar/go/strkey"
)

var padder = padding.NewPkcs7Padding(aes.BlockSize)

func getSeed(seed string) []byte {
	return strkey.MustDecode(strkey.VersionByteSeed, seed)
}

func encrypt(secret, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(secret, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	size := gcm.NonceSize()
	if len(ciphertext) < size {
		return nil, fmt.Errorf("ciphertext too short %d < %d ", len(ciphertext), size)
	}

	nonce, ciphertext := ciphertext[:size], ciphertext[size:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
