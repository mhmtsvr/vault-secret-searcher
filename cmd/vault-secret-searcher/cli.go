package main

import (
	search "vault-secret-searcher/pkg/cmd/search"
)

type CLI struct {
	Search search.Cmd `cmd:"" default:"withargs" help:"Recursively search vault secrets by key name (-k) or value content (-v)"`
	Debug  bool       `short:"d" help:"Enable debug logging"`
}
