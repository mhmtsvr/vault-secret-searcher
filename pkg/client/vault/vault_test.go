package vault

import (
	"testing"
)

func TestExtractData_KVv2(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"data": map[string]any{
				"username": "admin",
				"password": "secret123",
			},
			"metadata": map[string]any{
				"version": float64(1),
			},
		},
	}

	result := extractData(response)
	if result == nil {
		t.Fatal("expected data, got nil")
	}
	if result["username"] != "admin" {
		t.Errorf("expected username=admin, got %v", result["username"])
	}
	if result["password"] != "secret123" {
		t.Errorf("expected password=secret123, got %v", result["password"])
	}
}

func TestExtractData_KVv1(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"username": "admin",
			"password": "secret123",
		},
	}

	result := extractData(response)
	if result == nil {
		t.Fatal("expected data, got nil")
	}
	if result["username"] != "admin" {
		t.Errorf("expected username=admin, got %v", result["username"])
	}
}

func TestExtractData_NilData(t *testing.T) {
	response := map[string]any{
		"errors": []any{"permission denied"},
	}

	result := extractData(response)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestExtractData_EmptyResponse(t *testing.T) {
	response := map[string]any{}

	result := extractData(response)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}
