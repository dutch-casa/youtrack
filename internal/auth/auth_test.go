package auth

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreSaveLoadDelete(t *testing.T) {
	clearCredentialEnv(t)
	path := filepath.Join(t.TempDir(), "youtrack", "config.json")
	store := NewStore(path)
	creds := Credentials{BaseURL: "https://example.youtrack.cloud", Token: "perm:test"}

	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config mode = %o, want 0600", got)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.BaseURL != creds.BaseURL || loaded.Token != creds.Token {
		t.Fatalf("Load() = %#v, want %#v", loaded, creds)
	}

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Load() after delete error = %v, want ErrNotConfigured", err)
	}
}

func TestStoreSaveTightensExistingDirectory(t *testing.T) {
	clearCredentialEnv(t)
	dir := filepath.Join(t.TempDir(), "youtrack")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	path := filepath.Join(dir, "config.json")
	store := NewStore(path)

	if err := store.Save(Credentials{BaseURL: "https://example.youtrack.cloud", Token: "perm:test"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("config dir mode = %o, want 0700", got)
	}
}

func TestEnsureNonInteractiveMissingAuthIsActionable(t *testing.T) {
	clearCredentialEnv(t)
	store := NewStore(filepath.Join(t.TempDir(), "missing.json"))

	_, err := store.Ensure(strings.NewReader(""), &bytes.Buffer{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Ensure() error = %v, want ErrNotConfigured", err)
	}
	if !strings.Contains(err.Error(), "yt auth login") {
		t.Fatalf("Ensure() error = %q, want setup command", err.Error())
	}
	if !strings.Contains(err.Error(), EnvURL) || !strings.Contains(err.Error(), EnvToken) {
		t.Fatalf("Ensure() error = %q, want env var names", err.Error())
	}
}

func TestPromptSplitsURLAndToken(t *testing.T) {
	var out bytes.Buffer

	baseURL, err := PromptURL(strings.NewReader("https://example.youtrack.cloud\n"), &out)
	if err != nil {
		t.Fatalf("PromptURL() error = %v", err)
	}
	if baseURL != "https://example.youtrack.cloud" {
		t.Fatalf("PromptURL() = %q", baseURL)
	}

	out.Reset()
	token, err := PromptToken(strings.NewReader("perm:secret\n"), &out)
	if err != nil {
		t.Fatalf("PromptToken() error = %v", err)
	}
	if token != "perm:secret" {
		t.Fatalf("PromptToken() = %q", token)
	}
}

func TestCredentialsValidate(t *testing.T) {
	tests := []struct {
		name    string
		creds   Credentials
		wantErr bool
	}{
		{name: "valid", creds: Credentials{BaseURL: "https://example.youtrack.cloud", Token: "perm:test"}},
		{name: "missing url", creds: Credentials{Token: "perm:test"}, wantErr: true},
		{name: "relative url", creds: Credentials{BaseURL: "example", Token: "perm:test"}, wantErr: true},
		{name: "unsupported scheme", creds: Credentials{BaseURL: "ftp://example.youtrack.cloud", Token: "perm:test"}, wantErr: true},
		{name: "missing token", creds: Credentials{BaseURL: "https://example.youtrack.cloud"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.creds.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadEnvCredentialsValidates(t *testing.T) {
	t.Setenv(EnvURL, "not-absolute")
	t.Setenv(EnvToken, "perm:test")
	t.Setenv("YT_URL", "")
	t.Setenv("YT_TOKEN", "")

	_, err := NewStore(filepath.Join(t.TempDir(), "missing.json")).Load()
	if err == nil {
		t.Fatal("Load() error = nil, want invalid env credentials")
	}
}

func clearCredentialEnv(t *testing.T) {
	t.Helper()
	t.Setenv(EnvURL, "")
	t.Setenv(EnvToken, "")
	t.Setenv("YT_URL", "")
	t.Setenv("YT_TOKEN", "")
}
