package cmd

import (
	"errors"
	"net/http"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stellar/go/clients/horizon"
)

type handler func() error
type middleware func(handler) handler

func middlewares(fns ...middleware) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		h := func() error { return nil }
		for i := len(fns) - 1; i >= 0; i-- {
			h = fns[i](h)
		}

		return h()
	}
}

func checkDB(next handler) handler {
	return func() error {
		if viper.GetString("db") == "" {
			return errors.New("db path should be set")
		}

		return next()
	}
}

func checkSecret(next handler) handler {
	return func() error {
		if viper.GetString("secret") == "" {
			secret, err := promptPassword()
			if err != nil {
				return err
			}
			viper.Set("secret", secret)
		}

		return next()
	}
}

func promptPassword() (string, error) {
	prompt := promptui.Prompt{
		Label: "Password",
		Validate: func(input string) error {
			if len(input) < 8 {
				return errors.New("length be greater than 8")
			}
			return nil
		},
		Mask: '*',
	}
	return prompt.Run()
}

func getClient(testnet bool) *horizon.Client {
	client := horizon.DefaultPublicNetClient
	if testnet {
		client = horizon.DefaultTestNetClient
	}

	return client
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

func friendbotFund(addr string) {
	friendBotResp, err := http.Get("https://horizon-testnet.stellar.org/friendbot?addr=" + addr)
	if err != nil {
		fatal(err)
	}
	defer friendBotResp.Body.Close()
}
