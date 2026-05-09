# vault-secret-searcher

A CLI tool that recursively searches HashiCorp Vault KV secrets for matching key names or values.

## Install

```bash
go install ./cmd/vault-secret-searcher
```

Make sure `$GOPATH/bin` is on your `$PATH`. Add this to your `~/.zshrc` if it's not:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Usage

```bash
# Search key names (default) in secret/k8s/
vault-secret-searcher -s redis

# Search values instead of keys
vault-secret-searcher -v -s redis

# Search a different path
vault-secret-searcher -p secret/infra/ -s redis -v

# Login to Vault first, then search
vault-secret-searcher -l -s redis

# Enable debug logging
vault-secret-searcher -d -s redis
```

## Flags

| Flag | Description |
|------|-------------|
| `-s` | Search term (required, case-insensitive) |
| `-k` | Search by key names (default) |
| `-v` | Search by values |
| `-p` | Vault path (default: `secret/`) |
| `-l` | Run `vault login -method oidc` before searching |
| `-d` | Enable debug logging |

## Requirements

- `vault` CLI installed and configured
- Authenticated Vault session (or use `-l` to login)
