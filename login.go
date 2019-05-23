package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
	"golang.org/x/crypto/ssh/terminal"
)

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(color.Bold.Render("Enter Tidal Username: "))
	username, _ := reader.ReadString('\n')

	fmt.Print(color.Bold.Render("Enter Tidal Password: "))
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		return "", "", err
	}
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}
