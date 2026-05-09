package search

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

type mockClient struct {
	paths   map[string][]string
	secrets map[string]map[string]any
}

func (m *mockClient) Login() error { return nil }

func (m *mockClient) ListPath(path string) ([]string, error) {
	return m.paths[path], nil
}

func (m *mockClient) GetSecret(path string) (map[string]any, error) {
	return m.secrets[path], nil
}

func runCmd(t *testing.T, cmd *Cmd) string {
	t.Helper()
	var buf bytes.Buffer
	cmd.out = &buf
	if err := cmd.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestSearchKeys_CaseInsensitive(t *testing.T) {
	output := runCmd(t, &Cmd{
		Search: "redis",
		Path:   "secret/k8s/app1",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/k8s/app1": {
					"REDIS_URL":    "redis://localhost:6379",
					"DATABASE_URL": "postgres://localhost:5432",
				},
			},
		},
	})

	if !strings.Contains(output, "Found in: secret/k8s/app1") {
		t.Errorf("expected 'Found in' line, got:\n%s", output)
	}
	if !strings.Contains(output, "REDIS_URL") {
		t.Errorf("expected REDIS_URL in output, got:\n%s", output)
	}
	if strings.Contains(output, "DATABASE_URL") {
		t.Errorf("did not expect DATABASE_URL in output, got:\n%s", output)
	}
}

func TestSearchKeys_NoMatch(t *testing.T) {
	output := runCmd(t, &Cmd{
		Search: "redis",
		Path:   "secret/k8s/app1",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/k8s/app1": {
					"DATABASE_URL": "postgres://localhost:5432",
				},
			},
		},
	})

	if strings.Contains(output, "Found in:") {
		t.Errorf("did not expect 'Found in' line, got:\n%s", output)
	}
}

func TestSearchValues_CaseInsensitive(t *testing.T) {
	output := runCmd(t, &Cmd{
		Values: true,
		Search: "redis",
		Path:   "secret/k8s/app1",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/k8s/app1": {
					"cache_url":    "Redis://cache.example.com:6379",
					"DATABASE_URL": "postgres://localhost:5432",
				},
			},
		},
	})

	if !strings.Contains(output, "Found in: secret/k8s/app1") {
		t.Errorf("expected 'Found in' line, got:\n%s", output)
	}
	if !strings.Contains(output, "cache_url=Redis://cache.example.com:6379") {
		t.Errorf("expected cache_url entry in output, got:\n%s", output)
	}
	if strings.Contains(output, "DATABASE_URL") {
		t.Errorf("did not expect DATABASE_URL in output, got:\n%s", output)
	}
}

func TestWalkPath_Recursive(t *testing.T) {
	output := runCmd(t, &Cmd{
		Search: "redis",
		Path:   "secret/k8s/",
		client: &mockClient{
			paths: map[string][]string{
				"secret/k8s/":        {"team-a/", "app1"},
				"secret/k8s/team-a/": {"app2"},
			},
			secrets: map[string]map[string]any{
				"secret/k8s/app1":        {"redis_host": "redis-1.example.com"},
				"secret/k8s/team-a/app2": {"redis_host": "redis-2.example.com"},
			},
		},
	})

	if !strings.Contains(output, "Found in: secret/k8s/app1") {
		t.Errorf("expected match in app1, got:\n%s", output)
	}
	if !strings.Contains(output, "Found in: secret/k8s/team-a/app2") {
		t.Errorf("expected match in team-a/app2, got:\n%s", output)
	}
}

func TestWalkPath_LeafPath(t *testing.T) {
	output := runCmd(t, &Cmd{
		Search: "redis",
		Path:   "secret/k8s/single-app",
		client: &mockClient{
			paths: map[string][]string{},
			secrets: map[string]map[string]any{
				"secret/k8s/single-app": {"redis_password": "s3cret"},
			},
		},
	})

	if !strings.Contains(output, "Found in: secret/k8s/single-app") {
		t.Errorf("expected match for leaf path, got:\n%s", output)
	}
}

func TestSearchValues_NonStringValues(t *testing.T) {
	output := runCmd(t, &Cmd{
		Values: true,
		Search: "6379",
		Path:   "secret/k8s/app1",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/k8s/app1": {
					"redis_port":    float64(6379),
					"redis_enabled": true,
				},
			},
		},
	})

	if !strings.Contains(output, "redis_port=6379") {
		t.Errorf("expected redis_port=6379 in output, got:\n%s", output)
	}
}

func TestOutputFormat_Header(t *testing.T) {
	output := runCmd(t, &Cmd{
		Search: "redis",
		Path:   "secret/k8s/",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/k8s/": {},
			},
		},
	})

	if !strings.Contains(output, "Searching for: 'redis' in keys") {
		t.Errorf("expected header line, got:\n%s", output)
	}
	if !strings.Contains(output, "Vault path: secret/k8s/") {
		t.Errorf("expected path line, got:\n%s", output)
	}
	if !strings.Contains(output, "---") {
		t.Errorf("expected separator line, got:\n%s", output)
	}
}

func TestOutputFormat_HeaderValues(t *testing.T) {
	output := runCmd(t, &Cmd{
		Values: true,
		Search: "redis",
		Path:   "secret/k8s/",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/k8s/": {},
			},
		},
	})

	if !strings.Contains(output, "Searching for: 'redis' in values") {
		t.Errorf("expected 'in values' header line, got:\n%s", output)
	}
}

func TestMatchKeys_SortedOutput(t *testing.T) {
	output := runCmd(t, &Cmd{
		Search: "redis",
		Path:   "secret/app",
		client: &mockClient{
			secrets: map[string]map[string]any{
				"secret/app": {
					"redis_z_last":  "val",
					"redis_a_first": "val",
					"redis_m_mid":   "val",
				},
			},
		},
	})

	aIdx := strings.Index(output, "redis_a_first")
	mIdx := strings.Index(output, "redis_m_mid")
	zIdx := strings.Index(output, "redis_z_last")
	if aIdx > mIdx || mIdx > zIdx {
		t.Errorf("expected sorted output, got:\n%s", output)
	}
}
