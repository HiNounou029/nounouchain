package utils

import (
	"io/ioutil"
	"strings"
	"gopkg.in/urfave/cli.v1"
	"github.com/mattn/go-tty"
	"fmt"
	"os"
	"github.com/HiNounou029/polo-sdk-go/cmd/utils"
)

// MigrateFlags sets the global flag from a local flag when it's set.
// This is a temporary function used for migrating old command/flags to the
// new format.
//
// e.g. geth account new --keystore /tmp/mykeystore --lightkdf
//
// is equivalent after calling this method with:
//
// geth --keystore /tmp/mykeystore --lightkdf account new
//
// This allows the use of the existing configuration functionality.
// When all flags are migrated this function can be removed and the existing
// configuration functionality must be changed that is uses local flags
func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}

// MakePasswordList reads password lines from the file specified by the global --password flag.
func MakePasswordList(ctx *cli.Context) []string {
	path := ctx.GlobalString(utils.PasswordFileFlag.Name)
	if path == "" {
		return nil
	}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		utils.Fatalf("Failed to read password file: %v", err)
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}

// GetPassPhrase retrieves the password associated with an account, either fetched
// from a list of preloaded passphrases, or requested interactively from the user.
func GetPassPhrase(prompt string, confirmation bool, i int, passwords []string) string {
	// If a list of passwords was supplied, retrieve from them
	if len(passwords) > 0 {
		if i < len(passwords) {
			return passwords[i]
		}
		return passwords[len(passwords)-1]
	}
	// Otherwise prompt the user for the password
	if prompt != "" {
		fmt.Println(prompt)
	}
	password, err := readPasswordFromNewTTY("Passphrase: ")
	if err != nil {
		utils.Fatalf("Failed to read passphrase: %v", err)
	}
	if confirmation {
		confirm, err := readPasswordFromNewTTY("Repeat passphrase: ")
		if err != nil {
			utils.Fatalf("Failed to read passphrase confirmation: %v", err)
		}
		if password != confirm {
			utils.Fatalf("Passphrases do not match")
		}
	}
	return password
}

func readPasswordFromNewTTY(prompt string) (string, error) {
	t, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer t.Close()
	fmt.Fprint(t.Output(), prompt)
	pass, err := t.ReadPasswordNoEcho()
	if err != nil {
		return "", err
	}
	return pass, err
}

func CheckFileExistAndOverride(path string) bool {

	if Exists(path) {

		prompt := fmt.Sprintf("Override file[%s], yes or no ?\r\n", path)
		input, err := overrideFromNewTTy(prompt)
		if err != nil {
			utils.Fatalf("Failed to override file: %v", err)
		}

		if input == "yes" {
			return true
		}

		return false
	}

	return true
}

func overrideFromNewTTy(prompt string) (string, error) {
	t, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer t.Close()
	fmt.Fprint(t.Output(), prompt)
	input, err := t.ReadString()
	if err != nil {
		return "", err
	}
	return input, err
}


func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
