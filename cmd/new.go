// Copyright © 2018 Salim Alami Idrissi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/manifoldco/promptui"

	"golang.org/x/sync/errgroup"

	"github.com/celrenheit/alfred/wallet"

	"github.com/stellar/go/keypair"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "Command for creating new wallets",
	Long:    `Command for creating new wallets`,
	PreRunE: middlewares(checkDB, checkSecret),
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case len(args) > 1:
			fmt.Printf("too many arguments '%v'\n", args)
		case len(args) == 0 || args[0] == "wallet":
			path := viper.GetString("db")
			secret := viper.GetString("secret")
			m, err := wallet.OpenSecretString(path, secret)
			if err != nil {
				fatal("error opening backup:", err)
			}

			suffix := strings.ToUpper(viper.GetString("suffix"))
			kp, err := generateKP(suffix)
			if err != nil {
				fatal("error creating new wallet:", err)
			}

			name := cmd.Flag("name").Value.String()
			w := wallet.New(name, kp)
			err = m.AddWallet(w)
			if err != nil {
				fatal("error opening backup:", err)
			}

			// table := tablewriter.NewWriter(os.Stdout)
			// table.SetHeader([]string{"Address", "Seed"})
			// table.Append([]string{kp.Address(), kp.Seed()})
			// table.Render()

			if err := wallet.Write(path, m); err != nil {
				fatal("error opening backup:", err)
			}
		case args[0] == "contact":
			path := viper.GetString("db")
			secret := viper.GetString("secret")
			m, err := wallet.OpenSecretString(path, secret)
			if err != nil {
				fatal("error opening backup:", err)
			}

			name := cmd.Flag("name").Value.String()
			if name == "" {
				prompt := promptui.Prompt{
					Label: "Name of the contact",
					Validate: func(input string) error {
						if len(input) == 0 {
							return errors.New("should not be empty")
						}

						return nil
					},
				}
				name, err = prompt.Run()
				if err != nil {
					fatal(err)
				}
			}

			prompt := promptui.Prompt{
				Label: "Contact's address",
				Validate: func(input string) error {
					if !strings.HasPrefix(input, "G") {
						return errors.New("length be greater than 8")
					}

					return nil
				},
			}
			addr, err := prompt.Run()
			if err != nil {
				fatal(err)
			}

			memo, err := promptMemo()
			if err != nil {
				fatal(err)
			}

			err = m.AddContact(name, addr, memo)
			if err != nil {
				fatal("error opening backup:", err)
			}

			m.Unlock([]byte(secret))

			if err := wallet.Write(path, m); err != nil {
				fatal("error opening backup:", err)
			}
		default:
			fatalf("unknown argument '%s'\n", args[0])
		}
	},
}

func promptMemo() (memo *wallet.Memo, err error) {
	sl := promptui.Select{
		Label: "Memo",
		Items: []string{
			"None",
			"MEMO_TEXT",
			"MEMO_ID",
			"MEMO_HASH",
			"MEMO_RETURN",
		},
	}
	idx, _, err := sl.Run()
	if err != nil {
		return nil, err
	}

	if idx > 0 {
		kind := wallet.MemoKind(idx)
		prompt := promptui.Prompt{
			Label: "Memo",
			Validate: func(input string) error {
				_, err := wallet.MemoFromString(kind, input)
				return err
			},
		}

		got, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		memo, err = wallet.MemoFromString(kind, got)
	}

	return memo, err
}

func generateKP(suffix string) (*keypair.Full, error) {
	if suffix == "" {
		return keypair.Random()
	}
	start := time.Now()
	defer func() {
		fmt.Println("took", time.Since(start))
	}()

	var (
		baseCtx, cancel = context.WithCancel(context.Background())
		group, ctx      = errgroup.WithContext(baseCtx)
		ch              = make(chan *keypair.Full, 1)
	)

	for i := 0; i < runtime.NumCPU(); i++ {
		group.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				kp, err := keypair.Random()
				if err != nil {
					return err
				}

				if strings.HasSuffix(kp.Address(), suffix) {
					ch <- kp
					return nil
				}
			}
		})
	}

	errCh := make(chan error)

	go func() {
		errCh <- group.Wait()
	}()

	select {
	case err := <-errCh:
		return nil, err
	case kp := <-ch:
		cancel()
		<-errCh
		return kp, nil
	}
}

func init() {
	RootCmd.AddCommand(newCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	newCmd.Flags().String("name", "", "name of the wallet")
	newCmd.Flags().String("suffix", "", "suffix for vanity addresses")
	viper.BindPFlags(newCmd.Flags())
}
