package ytcli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpIsAvailableWithoutAuth(t *testing.T) {
	var out bytes.Buffer
	err := Execute(context.Background(), []string{"--help"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "Agent-friendly YouTrack CLI") {
		t.Fatalf("help output = %q", out.String())
	}
	for _, command := range []string{"projects", "users", "commands"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("help output missing %q: %q", command, out.String())
		}
	}
}

func TestAuthLoginStatusLogout(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	var out bytes.Buffer

	err := Execute(context.Background(), []string{
		"--config", config,
		"auth", "login",
		"--url", "https://example.youtrack.cloud",
		"--token", "perm:secret",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth login error = %v", err)
	}

	out.Reset()
	err = Execute(context.Background(), []string{"--config", config, "auth", "status"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth status error = %v", err)
	}
	if strings.Contains(out.String(), "perm:secret") {
		t.Fatalf("auth status leaked token: %q", out.String())
	}
	if !strings.Contains(out.String(), "configured") {
		t.Fatalf("auth status output = %q", out.String())
	}

	out.Reset()
	err = Execute(context.Background(), []string{"--config", config, "auth", "logout"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth logout error = %v", err)
	}
}

func TestMissingAuthNonInteractiveReturnsSetupError(t *testing.T) {
	config := filepath.Join(t.TempDir(), "missing.json")
	err := Execute(context.Background(), []string{"--config", config, "me"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want missing auth")
	}
	if !strings.Contains(err.Error(), "yt auth login") {
		t.Fatalf("Execute() error = %q, want setup command", err.Error())
	}
}
