package conformance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadJSON[T any](path string) (T, error) {
	var out T
	data, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("decode %s: %w", path, err)
	}
	return out, nil
}

func MustLoadJSON[T any](tb interface {
	Helper()
	Fatalf(format string, args ...any)
}, path string) T {
	tb.Helper()
	out, err := LoadJSON[T](path)
	if err != nil {
		tb.Fatalf("load json fixture %s: %v", path, err)
	}
	return out
}

func LoadCredentialValidationCases(dir string) ([]CredentialValidationCase, error) {
	return loadCaseList[CredentialValidationCase](filepath.Join(dir, "credential_cases.json"))
}

func LoadLoginPollCases(dir string) ([]LoginPollCase, error) {
	return loadCaseList[LoginPollCase](filepath.Join(dir, "login_poll_cases.json"))
}

func LoadInboundCases(dir string) ([]InboundCase, error) {
	return loadCaseList[InboundCase](filepath.Join(dir, "inbound_cases.json"))
}

func loadCaseList[T any](path string) ([]T, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return LoadJSON[[]T](path)
}
