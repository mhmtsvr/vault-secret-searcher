package search

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/log"

	"vault-secret-searcher/pkg/client/vault"
)

const (
	modeKeys   = "keys"
	modeValues = "values"
)

type Cmd struct {
	Keys   bool   `short:"k" xor:"mode" help:"Search by secret key names (default)"`
	Values bool   `short:"v" xor:"mode" help:"Search by secret values"`
	Search string `short:"s" required:"" help:"Term to search for (case-insensitive)"`
	Path   string `short:"p" default:"secret/" help:"Vault KV path to search recursively"`
	Login  bool   `short:"l" help:"Run vault OIDC login before searching"`

	client vault.Client
	out    io.Writer
}

func (c *Cmd) mode() string {
	if c.Values {
		return modeValues
	}
	return modeKeys
}

func (c *Cmd) Run(_ context.Context) error {
	if c.client == nil {
		c.client = vault.NewCLIClient()
	}
	if c.out == nil {
		c.out = os.Stdout
	}

	if c.Login {
		if err := c.client.Login(); err != nil {
			return err
		}
	}

	c.printf("Searching for: '%s' in %s\n", c.Search, c.mode())
	c.printf("Vault path: %s\n", c.Path)
	c.printf("---\n")

	return c.walkPath(c.Path)
}

func (c *Cmd) walkPath(path string) error {
	items, err := c.client.ListPath(path)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return c.searchSecret(path)
	}

	for _, item := range items {
		fullPath := path + item
		if strings.HasSuffix(item, "/") {
			if err := c.walkPath(fullPath); err != nil {
				return err
			}
		} else {
			if err := c.searchSecret(fullPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cmd) searchSecret(path string) error {
	data, err := c.client.GetSecret(path)
	if err != nil {
		return err
	}

	if data == nil {
		log.Debug("No data", "path", path)
		return nil
	}

	if c.mode() == modeKeys {
		c.matchKeys(path, data)
	} else {
		c.matchValues(path, data)
	}

	return nil
}

func (c *Cmd) matchKeys(path string, data map[string]any) {
	term := strings.ToLower(c.Search)
	var matches []string

	for key := range data {
		if strings.Contains(strings.ToLower(key), term) {
			matches = append(matches, key)
		}
	}

	sort.Strings(matches)

	if len(matches) > 0 {
		c.printf("Found in: %s\n", path)
		c.printf("  Matching keys:\n")
		for _, m := range matches {
			c.printf("    %s\n", m)
		}
	}
}

func (c *Cmd) matchValues(path string, data map[string]any) {
	term := strings.ToLower(c.Search)
	var matches []string

	for key, val := range data {
		entry := fmt.Sprintf("%s=%v", key, val)
		if strings.Contains(strings.ToLower(entry), term) {
			matches = append(matches, entry)
		}
	}

	sort.Strings(matches)

	if len(matches) > 0 {
		c.printf("Found in: %s\n", path)
		c.printf("  Matching entries:\n")
		for _, m := range matches {
			c.printf("    %s\n", m)
		}
	}
}

func (c *Cmd) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(c.out, format, args...)
}
