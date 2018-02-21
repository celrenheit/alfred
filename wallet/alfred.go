package wallet

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/stellar/go/keypair"
)

type Alfred struct {
	Stellar WalletsManager `yaml:"stellar,omitempty"`
	secret  []byte
}

func (a *Alfred) Unlock(secret []byte) error {
	if secret == nil {
		return nil
	}

	h := sha256.New()
	h.Write([]byte(secret))
	a.secret = h.Sum(nil)
	return nil
}

func (a *Alfred) IsUnlocked() bool { return a.secret != nil }

type alfredyaml struct {
	Stellar struct {
		Wallets  []walletyaml       `yaml:"wallets,omitempty"`
		Contacts map[string]Contact `yaml:"contacts,omitempty"`
	} `yaml:"stellar,omitempty"`
}

type walletyaml struct {
	Name    string `yaml:"name,omitempty"`
	Address string `yaml:"address,omitempty"`
	Seed    string `yaml:"seed,omitempty"`
}

func (a Alfred) MarshalYAML() (interface{}, error) {
	if a.secret == nil {
		return nil, errors.New("no secret set")
	}

	j := alfredyaml{}
	j.Stellar.Contacts = a.Stellar.Contacts
	for _, w := range a.Stellar.Wallets {
		kp, ok := w.Keypair.(*keypair.Full)
		if !ok || a.secret == nil {
			return nil, errors.New("you should unlock alfred for writing")
		}
		seed := getSeed(kp.Seed())
		encrypted, err := encrypt(a.secret, seed)
		if err != nil {
			return nil, err
		}

		encoded := base64.RawStdEncoding.EncodeToString(encrypted)
		j.Stellar.Wallets = append(j.Stellar.Wallets, walletyaml{
			Name:    w.Name,
			Address: kp.Address(),
			Seed:    encoded,
		})
	}

	return j, nil
}

func (a *Alfred) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aj alfredyaml

	if err := unmarshal(&aj); err != nil {
		return err
	}

	for _, j := range aj.Stellar.Wallets {
		w := &Wallet{}
		w.Name = j.Name
		if a.secret == nil {
			var err error
			w.Keypair, err = keypair.Parse(j.Address)
			if err != nil {
				return err
			}
		} else {
			decoded, err := base64.RawStdEncoding.DecodeString(j.Seed)
			if err != nil {
				return err
			}

			seed, err := decrypt(a.secret, decoded)
			if err != nil {
				return err
			}

			var seed32 [32]byte
			copy(seed32[:], seed)
			kp, err := keypair.FromRawSeed(seed32)
			if err != nil {
				return err
			}

			if kp.Address() != j.Address {
				return errors.New("address mismatch, password may be incorrect")
			}

			w.Name = j.Name
			w.Keypair = kp
		}

		a.Stellar.Wallets = append(a.Stellar.Wallets, w)
	}
	a.Stellar.Contacts = aj.Stellar.Contacts
	return nil
}

type WalletsManager struct {
	Wallets  []*Wallet          `yaml:"wallets,omitempty"`
	Contacts map[string]Contact `yaml:"contacts,omitempty"`
}

type Contact struct {
	Address string `yaml:"address,omitempty"`
	Memo    *Memo  `yaml:"memo,omitempty"`
}

func OpenSecretString(path string, secretStr string) (*Alfred, error) {
	var secret []byte
	if len(secretStr) > 0 {
		secret = []byte(secretStr)
	}

	return Open(path, secret)
}

func Open(path string, secret []byte) (*Alfred, error) {
	b, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {

	} else if err != nil {
		return nil, err
	}

	var a Alfred
	a.Unlock(secret)
	if len(b) > 0 {
		err = yaml.Unmarshal(b, &a)
		if err != nil {
			return nil, err
		}
	}
	return &a, nil
}

func Write(path string, m *Alfred) error {
	ciphertext, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, ciphertext, 0600)
}

func (m *Alfred) AddWallet(w *Wallet) error {
	if m.WalletByAddress(w.Keypair.Address()) != nil {
		return errors.New("wallet already exists")
	}

	m.Stellar.Wallets = append(m.Stellar.Wallets, w)
	return nil
}

func (m *Alfred) AddContact(name, addr string, memo *Memo) error {
	if m.Stellar.Contacts == nil {
		m.Stellar.Contacts = map[string]Contact{}
	}

	if _, ok := m.Stellar.Contacts[name]; ok {
		return errors.New("contact already exists")
	}

	if _, err := keypair.Parse(addr); err != nil {
		return err
	}

	m.Stellar.Contacts[name] = Contact{
		Address: addr,
		Memo:    memo,
	}

	return nil
}

func (m *Alfred) WalletByAddress(address string) *Wallet {
	for _, w := range m.Stellar.Wallets {
		if w.Keypair.Address() == address {
			return w
		}
	}

	return nil
}

func (m *Alfred) WalletByName(name string) *Wallet {
	for _, w := range m.Stellar.Wallets {
		if w.Name == name {
			return w
		}
	}

	return nil
}
