package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
)

type Client interface {
	Login() error
	ListPath(path string) ([]string, error)
	GetSecret(path string) (map[string]any, error)
}

type CLIClient struct{}

func NewCLIClient() *CLIClient {
	return &CLIClient{}
}

func (c *CLIClient) Login() error {
	cmd := exec.Command("vault", "login", "-method", "oidc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vault OIDC login failed: %w", err)
	}

	return nil
}

func (c *CLIClient) ListPath(path string) ([]string, error) {
	log.Debug("Listing", "path", path)

	cmd := exec.Command("vault", "kv", "list", "-format=json", path) //nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug("List failed", "path", path)
		return nil, nil
	}

	var items []string
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("parsing list output for %s: %w", path, err)
	}

	return items, nil
}

func (c *CLIClient) GetSecret(path string) (map[string]any, error) {
	log.Debug("Getting secret", "path", path)

	cmd := exec.Command("vault", "kv", "get", "-format=json", path) //nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		errOut := string(output)
		if strings.Contains(errOut, "permission denied") {
			return nil, fmt.Errorf("permission denied at %s — have you logged in?", path)
		}
		if strings.Contains(errOut, "no such host") {
			return nil, fmt.Errorf("can't reach vault — are you on vpn?")
		}
		log.Debug("No data", "path", path)
		return nil, nil
	}

	var response map[string]any
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("parsing secret at %s: %w", path, err)
	}

	return extractData(response), nil
}

func extractData(response map[string]any) map[string]any {
	data, ok := response["data"].(map[string]any)
	if !ok {
		return nil
	}

	// KV v2: nested .data.data
	if inner, ok := data["data"].(map[string]any); ok {
		return inner
	}

	return data
}
