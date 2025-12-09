package pepper

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "pzsp-teams-cache"
	userName    = "pepper"
)

func GetOrAskPepper() (string, error) {
	value, err := keyring.Get(serviceName, userName)
	if err == nil && strings.TrimSpace(value) != "" {
		return value, nil
	}
	fmt.Print("Enter secret pepper: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading pepper: %w", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("pepper cannot be empty")
	}
	if err := keyring.Set(serviceName, userName, input); err != nil {
		return "", fmt.Errorf("storing pepper in keyring: %w", err)
	}

	return input, nil
}
