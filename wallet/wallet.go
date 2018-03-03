package wallet

import (
	"fmt"
	"net/http"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

type Wallet struct {
	Name    string
	Keypair keypair.KP
}

func (w *Wallet) String() string {
	str := TrimAddress(w.Keypair.Address())
	if w.Name != w.Keypair.Address() {
		str = fmt.Sprintf("%s (%s)", w.Name, str)
	}
	return str
}

func New(name string, keypair *keypair.Full) *Wallet {
	if name == "" {
		name = keypair.Address()
	}
	return &Wallet{
		Name:    name,
		Keypair: keypair,
	}
}

func (w *Wallet) Balances(testnet bool) ([]horizon.Balance, error) {
	client := getClient(testnet)
	account, _, err := getAccount(client, w.Keypair.Address())
	if err != nil {
		return nil, err
	}

	if len(account.Balances) == 0 {
		account.Balances = append(account.Balances, horizon.Balance{
			Balance: "0",
			Asset: horizon.Asset{
				Type: "native",
			},
		})
	}

	return account.Balances, nil
}

func getAccount(client *horizon.Client, account string) (horizon.Account, bool, error) {
	hAccount, err := client.LoadAccount(account)
	if err != nil {
		if err, ok := err.(*horizon.Error); ok && err.Response.StatusCode == http.StatusNotFound {
			return hAccount, false, nil
		}
		return hAccount, false, err
	}

	return hAccount, true, nil
}

func getClient(testnet bool) *horizon.Client {
	client := horizon.DefaultPublicNetClient
	if testnet {
		client = horizon.DefaultTestNetClient
	}

	return client
}
