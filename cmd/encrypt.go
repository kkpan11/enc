package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/spf13/cobra"
)

// Keys:
// + password
// + key (name, path, binary, or text)
// + key with passphrase

type Encrypt struct {
	cfg      Config
	password string
	key      string
}

func (e Encrypt) Command() *cobra.Command {
	c := &cobra.Command{
		Use:     "encrypt",
		Aliases: []string{"encode", "e"},
		Args:    cobra.NoArgs,
		Short:   "Encrypt the message",
		RunE: func(_ *cobra.Command, args []string) error {
			return e.run()
		},
	}
	c.Flags().StringVarP(&e.password, "password", "p", "", "password to use")
	c.Flags().StringVarP(&e.key, "key", "k", "", "path to the key to use")
	return c
}

func (cmd Encrypt) run() error {
	if !cmd.cfg.HasStdin() {
		return errors.New("no file passed into stdin")
	}
	data, err := io.ReadAll(cmd.cfg)
	if err != nil {
		return fmt.Errorf("cannot read from stdin: %v", err)
	}
	message := crypto.NewPlainMessage(data)
	var encrypted *crypto.PGPMessage
	if cmd.key != "" {
		key, err := ReadKeyFile(cmd.key)
		if err != nil {
			return fmt.Errorf("cannot read key: %v", err)
		}
		if cmd.password != "" {
			key, err = key.Unlock([]byte(cmd.password))
			if err != nil {
				return fmt.Errorf("cannot unlock key: %v", err)
			}
		}
		keyring, err := crypto.NewKeyRing(key)
		if err != nil {
			return fmt.Errorf("cannot create keyring: %v", err)
		}
		encrypted, err = keyring.Encrypt(message, nil)
	} else if cmd.password != "" {
		encrypted, err = crypto.EncryptMessageWithPassword(message, []byte(cmd.password))
	} else {
		return errors.New("a password or a key required")
	}
	if err != nil {
		return fmt.Errorf("cannot encrypt the message: %v", err)
	}
	_, err = cmd.cfg.Write(encrypted.GetBinary())
	return err
}
