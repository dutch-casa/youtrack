package ytcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
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
	for _, command := range []string{"capabilities", "projects", "users", "articles", "agiles", "helpdesk", "commands", "attachments", "activities", "links", "upgrade"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("help output missing %q: %q", command, out.String())
		}
	}
	if !strings.Contains(out.String(), "work-items") {
		t.Fatalf("help output missing work-items: %q", out.String())
	}
}

func TestCapabilitiesIsAvailableWithoutAuth(t *testing.T) {
	config := filepath.Join(t.TempDir(), "missing.json")
	var out bytes.Buffer
	err := Execute(context.Background(), []string{"--config", config, "capabilities"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("capabilities error = %v", err)
	}

	var doc struct {
		SchemaVersion    int      `json:"schemaVersion"`
		DefaultMode      string   `json:"defaultMode"`
		DefaultOutput    string   `json:"defaultOutput"`
		RequiresAuth     bool     `json:"requiresAuth"`
		CommandNames     []string `json:"commandNames"`
		CommandReference []struct {
			Command      string `json:"command"`
			AuthRequired bool   `json:"authRequired"`
			Mutates      bool   `json:"mutates"`
			Output       string `json:"output"`
			Flags        []struct {
				Name      string `json:"name"`
				Required  bool   `json:"required"`
				Repeat    bool   `json:"repeat"`
				Exclusive string `json:"exclusiveWith"`
			} `json:"flags"`
		} `json:"commandReference"`
		Completeness []struct {
			Command  string   `json:"command"`
			Coverage string   `json:"coverage"`
			UseFor   []string `json:"useFor"`
		} `json:"completeness"`
		Interactive struct {
			Command  string   `json:"command"`
			Sections []string `json:"sections"`
		} `json:"interactive"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("capabilities output is not JSON: %v; output %q", err, out.String())
	}
	if doc.SchemaVersion != 1 || doc.DefaultMode != "non-interactive" || doc.DefaultOutput != "json" || doc.RequiresAuth {
		t.Fatalf("capabilities = %#v, want auth-free non-interactive json schema v1", doc)
	}
	if !containsString(doc.CommandNames, "yt") || !containsString(doc.CommandNames, "youtrack") {
		t.Fatalf("command names = %#v, want yt and youtrack", doc.CommandNames)
	}
	createIssue := findCapabilityCommand(doc.CommandReference, "yt issues create")
	if createIssue.Command == "" || !createIssue.AuthRequired || !createIssue.Mutates {
		t.Fatalf("yt issues create capability = %#v, want authenticated mutating command", createIssue)
	}
	if !findCapabilityFlag(createIssue.Flags, "project").Required || !findCapabilityFlag(createIssue.Flags, "summary").Required {
		t.Fatalf("yt issues create flags = %#v, want required project and summary", createIssue.Flags)
	}
	if findCapabilityFlag(createIssue.Flags, "description-file").Exclusive != "description,description-stdin" {
		t.Fatalf("description-file flag = %#v, want text source exclusivity", findCapabilityFlag(createIssue.Flags, "description-file"))
	}
	raw := findCapabilityCommand(doc.CommandReference, "yt raw PATH")
	if raw.Output != "Exact response bytes to stdout or --output-file." || !findCapabilityFlag(raw.Flags, "header").Repeat || !findCapabilityFlag(raw.Flags, "query").Repeat {
		t.Fatalf("yt raw capability = %#v, want exact byte output and repeatable header/query", raw)
	}
	if !capabilityBridgeContains(doc.Completeness, "yt commands apply", "YouTrack command-language workflows") {
		t.Fatalf("capabilities completeness = %#v, want command-language bridge", doc.Completeness)
	}
	if !capabilityBridgeContains(doc.Completeness, "yt raw", "Any YouTrack REST endpoint") {
		t.Fatalf("capabilities completeness = %#v, want raw REST bridge", doc.Completeness)
	}
	if doc.Interactive.Command != "yt interactive" || !containsString(doc.Interactive.Sections, "knowledge base") {
		t.Fatalf("interactive capabilities = %#v, want knowledge base section", doc.Interactive)
	}
}

func findCapabilityCommand(commands []struct {
	Command      string `json:"command"`
	AuthRequired bool   `json:"authRequired"`
	Mutates      bool   `json:"mutates"`
	Output       string `json:"output"`
	Flags        []struct {
		Name      string `json:"name"`
		Required  bool   `json:"required"`
		Repeat    bool   `json:"repeat"`
		Exclusive string `json:"exclusiveWith"`
	} `json:"flags"`
}, command string) struct {
	Command      string `json:"command"`
	AuthRequired bool   `json:"authRequired"`
	Mutates      bool   `json:"mutates"`
	Output       string `json:"output"`
	Flags        []struct {
		Name      string `json:"name"`
		Required  bool   `json:"required"`
		Repeat    bool   `json:"repeat"`
		Exclusive string `json:"exclusiveWith"`
	} `json:"flags"`
} {
	for _, candidate := range commands {
		if candidate.Command == command {
			return candidate
		}
	}
	return struct {
		Command      string `json:"command"`
		AuthRequired bool   `json:"authRequired"`
		Mutates      bool   `json:"mutates"`
		Output       string `json:"output"`
		Flags        []struct {
			Name      string `json:"name"`
			Required  bool   `json:"required"`
			Repeat    bool   `json:"repeat"`
			Exclusive string `json:"exclusiveWith"`
		} `json:"flags"`
	}{}
}

func findCapabilityFlag(flags []struct {
	Name      string `json:"name"`
	Required  bool   `json:"required"`
	Repeat    bool   `json:"repeat"`
	Exclusive string `json:"exclusiveWith"`
}, name string) struct {
	Name      string `json:"name"`
	Required  bool   `json:"required"`
	Repeat    bool   `json:"repeat"`
	Exclusive string `json:"exclusiveWith"`
} {
	for _, flag := range flags {
		if flag.Name == name {
			return flag
		}
	}
	return struct {
		Name      string `json:"name"`
		Required  bool   `json:"required"`
		Repeat    bool   `json:"repeat"`
		Exclusive string `json:"exclusiveWith"`
	}{}
}

func capabilityBridgeContains(bridges []struct {
	Command  string   `json:"command"`
	Coverage string   `json:"coverage"`
	UseFor   []string `json:"useFor"`
}, command, coverage string) bool {
	for _, bridge := range bridges {
		if strings.Contains(bridge.Command, command) && strings.Contains(bridge.Coverage, coverage) {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestAuthLoginStatusLogout(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	var out bytes.Buffer

	err := Execute(context.Background(), []string{
		"--config", config,
		"auth", "login",
		"--no-verify",
		"--url", "https://example.youtrack.cloud",
		"--token", "perm:secret",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth login error = %v", err)
	}
	var loginResult struct {
		Saved      bool   `json:"saved"`
		ConfigPath string `json:"configPath"`
	}
	if err := json.Unmarshal(out.Bytes(), &loginResult); err != nil {
		t.Fatalf("auth login output is not JSON: %v; output %q", err, out.String())
	}
	if !loginResult.Saved || loginResult.ConfigPath != config {
		t.Fatalf("auth login output = %#v, want saved config path", loginResult)
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
	var logoutResult struct {
		Removed bool `json:"removed"`
	}
	if err := json.Unmarshal(out.Bytes(), &logoutResult); err != nil {
		t.Fatalf("auth logout output is not JSON: %v; output %q", err, out.String())
	}
	if !logoutResult.Removed {
		t.Fatalf("auth logout output = %#v, want removed true", logoutResult)
	}
}

func TestAuthLoginOpenStartsBrowserSetup(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	var out bytes.Buffer
	var errOut bytes.Buffer
	var opened []string
	previousOpenBrowser := openBrowser
	openBrowser = func(rawURL string) error {
		opened = append(opened, rawURL)
		return nil
	}
	t.Cleanup(func() {
		openBrowser = previousOpenBrowser
	})

	err := Execute(context.Background(), []string{
		"--config", config,
		"auth", "login",
		"--open",
		"--no-verify",
		"--url", "https://example.youtrack.cloud/",
		"--token", "perm:secret",
	}, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("auth login --open error = %v", err)
	}
	if len(opened) != 1 || opened[0] != "https://example.youtrack.cloud/users/me?tab=account-security" {
		t.Fatalf("opened = %#v, want account security URL", opened)
	}
	if !strings.Contains(errOut.String(), "Account Security") {
		t.Fatalf("stderr = %q, want token setup guidance", errOut.String())
	}
	if strings.Contains(errOut.String(), "perm:secret") || strings.Contains(out.String(), "perm:secret") {
		t.Fatalf("auth login --open leaked token; stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

func TestTokenSetupURLPreservesSelfHostedBasePath(t *testing.T) {
	got, err := tokenSetupURL("https://youtrack.example.com/youtrack/")
	if err != nil {
		t.Fatalf("tokenSetupURL() error = %v", err)
	}
	want := "https://youtrack.example.com/youtrack/users/me?tab=account-security"
	if got != want {
		t.Fatalf("tokenSetupURL() = %q, want %q", got, want)
	}
}

func TestAuthLoginVerifiesBeforeSaving(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users/me" {
			t.Fatalf("request path = %q, want /api/users/me", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer perm:secret" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte(`{"id":"u-1","login":"jane"}`))
	}))
	t.Cleanup(server.Close)

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", config,
		"auth", "login",
		"--url", server.URL,
		"--token", "perm:secret",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth login error = %v", err)
	}

	var result struct {
		Saved    bool   `json:"saved"`
		Verified bool   `json:"verified"`
		User     string `json:"user"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("auth login output is not JSON: %v; output %q", err, out.String())
	}
	if !result.Saved || !result.Verified || result.User != "jane" {
		t.Fatalf("auth login output = %#v, want verified jane", result)
	}
}

func TestAuthLoginDoesNotSaveFailedVerification(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	err := Execute(context.Background(), []string{
		"--config", config,
		"auth", "login",
		"--url", server.URL,
		"--token", "perm:bad",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("auth login error = nil, want verification error")
	}
	if !strings.Contains(err.Error(), "verify credentials") {
		t.Fatalf("auth login error = %q, want verification context", err.Error())
	}

	var out bytes.Buffer
	if err := Execute(context.Background(), []string{"--config", config, "auth", "status"}, strings.NewReader(""), &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("auth status error = %v", err)
	}
	if !strings.Contains(out.String(), `"configured": false`) {
		t.Fatalf("auth status output = %q, want unconfigured", out.String())
	}
}

func TestUpgradeRunsDownloadedInstallerForCurrentBinary(t *testing.T) {
	const installerScript = "echo installer\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(installerScript))
	}))
	t.Cleanup(server.Close)

	var installed struct {
		script string
		binDir string
		name   string
	}
	previousRunInstallScript := runInstallScript
	runInstallScript = func(ctx context.Context, script []byte, binDir, name string, out, errOut io.Writer) error {
		installed.script = string(script)
		installed.binDir = binDir
		installed.name = name
		_, _ = fmt.Fprintln(out, "installer progress")
		return nil
	}
	t.Cleanup(func() {
		runInstallScript = previousRunInstallScript
	})

	targetDir := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := Execute(context.Background(), []string{
		"upgrade",
		"--installer-url", server.URL,
		"--bin-dir", targetDir,
		"--name", "yt-test",
	}, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("upgrade error = %v", err)
	}
	if installed.script != installerScript || installed.binDir != targetDir || installed.name != "yt-test" {
		t.Fatalf("installer invocation = %#v", installed)
	}
	var result struct {
		Updated bool   `json:"updated"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("upgrade output is not JSON: %v; output %q", err, out.String())
	}
	if strings.Contains(out.String(), "installer progress") {
		t.Fatalf("upgrade stdout included installer progress: %q", out.String())
	}
	if !strings.Contains(errOut.String(), "installer progress") {
		t.Fatalf("upgrade stderr = %q, want installer progress", errOut.String())
	}
	if !result.Updated || result.Path != filepath.Join(targetDir, "yt-test") {
		t.Fatalf("upgrade output = %#v", result)
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

func TestMissingAuthPromptVerifiesBeforeSaving(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/api/users/me" {
			t.Fatalf("request path = %q, want /api/users/me", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer perm:secret" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte(`{"id":"u-1","login":"jane"}`))
	}))
	t.Cleanup(server.Close)

	previousCanPrompt := canPrompt
	canPrompt = func(in io.Reader) bool {
		return true
	}
	t.Cleanup(func() {
		canPrompt = previousCanPrompt
	})

	var out bytes.Buffer
	input := strings.NewReader(server.URL + "\nperm:secret\n")
	err := Execute(context.Background(), []string{"--config", config, "me"}, input, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want verification plus command request", requests)
	}
	if !strings.Contains(out.String(), "jane") {
		t.Fatalf("me output = %q, want user", out.String())
	}

	out.Reset()
	if err := Execute(context.Background(), []string{"--config", config, "auth", "status"}, strings.NewReader(""), &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("auth status error = %v", err)
	}
	if !strings.Contains(out.String(), `"configured": true`) {
		t.Fatalf("auth status output = %q, want configured", out.String())
	}
}

func TestMissingAuthPromptDoesNotSaveFailedVerification(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	previousCanPrompt := canPrompt
	canPrompt = func(in io.Reader) bool {
		return true
	}
	t.Cleanup(func() {
		canPrompt = previousCanPrompt
	})

	input := strings.NewReader(server.URL + "\nperm:bad\n")
	err := Execute(context.Background(), []string{"--config", config, "me"}, input, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want verification error")
	}
	if !strings.Contains(err.Error(), "verify credentials") {
		t.Fatalf("Execute() error = %q, want verification context", err.Error())
	}

	var out bytes.Buffer
	if err := Execute(context.Background(), []string{"--config", config, "auth", "status"}, strings.NewReader(""), &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("auth status error = %v", err)
	}
	if !strings.Contains(out.String(), `"configured": false`) {
		t.Fatalf("auth status output = %q, want unconfigured", out.String())
	}
}

func TestIssuesUpdateRequiresChangedField(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	err := Execute(context.Background(), []string{"--config", config, "issues", "update", "ABC-1"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "--summary") {
		t.Fatalf("Execute() error = %q, want field guidance", err.Error())
	}
}

func TestListPaginationRejectsInvalidValuesBeforeAuth(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "top",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "issues", "list", "--top", "0"},
			want: "--top",
		},
		{
			name: "skip",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "activities", "list", "ABC-1", "--skip", "-1"},
			want: "--skip",
		},
		{
			name: "interactive top",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "interactive", "--top", "0"},
			want: "--top",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Execute(context.Background(), tt.args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil {
				t.Fatal("Execute() error = nil, want pagination validation")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), tt.want)
			}
			if strings.Contains(err.Error(), "yt auth login") {
				t.Fatalf("Execute() error = %q, validated pagination after auth", err.Error())
			}
		})
	}
}

func TestCommentAddReadsTextFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/comments" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"c-1","text":"from stdin","author":{"login":"jane"},"created":1}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"comments", "add", "ABC-1", "--text-stdin",
	}, strings.NewReader("from stdin"), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "from stdin") {
		t.Fatalf("output = %q, want stdin comment", out.String())
	}
}

func TestWorkItemsAddRequiresPositiveMinutes(t *testing.T) {
	err := Execute(context.Background(), []string{"work-items", "add", "ABC-1"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want minutes validation")
	}
	if !strings.Contains(err.Error(), "--minutes") {
		t.Fatalf("Execute() error = %q, want minutes guidance", err.Error())
	}
}

func TestAttachmentsAddRequiresFile(t *testing.T) {
	err := Execute(context.Background(), []string{"attachments", "add", "ABC-1"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want file validation")
	}
	if !strings.Contains(err.Error(), "--file") {
		t.Fatalf("Execute() error = %q, want file guidance", err.Error())
	}
}

func TestAttachmentsAddUploadsFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/attachments" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart form", r.Header.Get("Content-Type"))
		}
		file, header, err := r.FormFile("upload")
		if err != nil {
			t.Fatalf("upload field: %v", err)
		}
		defer file.Close()
		if header.Filename != "evidence.txt" {
			t.Fatalf("filename = %q", header.Filename)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(file); err != nil {
			t.Fatalf("read upload: %v", err)
		}
		if body.String() != "evidence" {
			t.Fatalf("upload body = %q", body.String())
		}
		_, _ = w.Write([]byte(`[{"id":"134-1","name":"evidence.txt","size":8}]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	path := filepath.Join(t.TempDir(), "evidence.txt")
	if err := os.WriteFile(path, []byte("evidence"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"attachments", "add", "ABC-1", "--file", path,
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "evidence.txt") {
		t.Fatalf("output = %q, want attachment name", out.String())
	}
}

func TestActivitiesListUsesDefaultCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/activities" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		categories := r.URL.Query().Get("categories")
		if !strings.Contains(categories, "CommentsCategory") || !strings.Contains(categories, "CustomFieldCategory") {
			t.Fatalf("categories = %q, want default activity categories", categories)
		}
		if got := r.URL.Query().Get("reverse"); got != "true" {
			t.Fatalf("reverse = %q, want true", got)
		}
		_, _ = w.Write([]byte(`[{"id":"a-1","$type":"CommentActivityItem","timestamp":1,"author":{"login":"jane"},"target":{"text":"Looks fixed"}}]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"history", "list", "ABC-1",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "Looks fixed") {
		t.Fatalf("output = %q, want activity target", out.String())
	}
}

func TestActivitiesListNarrowsCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("categories"); got != "CommentsCategory" {
			t.Fatalf("categories = %q, want explicit category only", got)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"activities", "list", "ABC-1", "--category", "CommentsCategory",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestLinksList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/links" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("$top"); got != "5" {
			t.Fatalf("$top = %q, want 5", got)
		}
		_, _ = w.Write([]byte(`[{"id":"80-1s","direction":"OUTWARD","linkType":{"name":"Depend","sourceToTarget":"is required for","targetToSource":"depends on","directed":true},"issues":[{"id":"2-44","idReadable":"ABC-2","summary":"Linked issue"}]}]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"--format", "table",
		"links", "list", "ABC-1", "--top", "5",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "ABC-2") || !strings.Contains(out.String(), "is required for") {
		t.Fatalf("output = %q, want linked issue and direction", out.String())
	}
}

func TestRawReadsBodyFromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(r.Body); err != nil {
			t.Fatalf("read body: %v", err)
		}
		if body.String() != `{"summary":"from file"}` {
			t.Fatalf("body = %q", body.String())
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	path := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(path, []byte(`{"summary":"from file"}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "--method", "POST", "--body-file", path,
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.String() != `{"ok":true}` {
		t.Fatalf("output = %q, want exact raw response", out.String())
	}
}

func TestRawReadsBodyFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(r.Body); err != nil {
			t.Fatalf("read body: %v", err)
		}
		if body.String() != `{"summary":"from stdin"}` {
			t.Fatalf("body = %q", body.String())
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "-X", "POST", "--body-stdin",
	}, strings.NewReader(`{"summary":"from stdin"}`), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawSendsCustomContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "text/plain" {
			t.Fatalf("Content-Type = %q, want text/plain", got)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(r.Body); err != nil {
			t.Fatalf("read body: %v", err)
		}
		if body.String() != "plain text" {
			t.Fatalf("body = %q, want plain text", body.String())
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/custom", "-X", "POST", "--content-type", "text/plain", "--body", "plain text",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawSendsCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer perm:test" {
			t.Fatalf("Authorization = %q, want configured bearer token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/xml" {
			t.Fatalf("Accept = %q, want custom accept header", got)
		}
		if got := r.Header.Get("X-Youtrack-Trace"); got != "agent-run-1" {
			t.Fatalf("X-YouTrack-Trace = %q, want agent-run-1", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/custom", "-H", "Accept: application/xml", "--header", "X-YouTrack-Trace: agent-run-1",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawSendsQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query()["fields"]; len(got) != 1 || got[0] != "id,idReadable" {
			t.Fatalf("fields query = %#v, want id,idReadable", got)
		}
		if got := r.URL.Query()["tag"]; len(got) != 2 || got[0] != "agent" || got[1] != "urgent" {
			t.Fatalf("tag query = %#v, want repeated values", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "--query", "fields=id,idReadable", "-q", "tag=agent", "-q", "tag=urgent",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawWritesOutputFile(t *testing.T) {
	want := []byte{0x00, 0x01, 0x02, 'Y', 'T', '\n'}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(want)
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	outputPath := filepath.Join(t.TempDir(), "raw.bin")
	if err := os.WriteFile(outputPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed output file: %v", err)
	}
	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/download", "--output-file", outputPath,
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty when output file is set", out.String())
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("output file = %v, want exact raw bytes %v", got, want)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat output file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("output file mode = %v, want 0600", got)
	}
}

func TestRawWritesExactBytesToStdout(t *testing.T) {
	want := []byte{0x00, 0x01, 0x02, 'Y', 'T'}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(want)
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/download",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("stdout = %v, want exact raw bytes %v", out.Bytes(), want)
	}
}

func TestRawRejectsInvalidQueryBeforeAuth(t *testing.T) {
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "--query", "fields",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want query validation")
	}
	if !strings.Contains(err.Error(), "name=value") {
		t.Fatalf("Execute() error = %q, want query shape guidance", err.Error())
	}
	if strings.Contains(err.Error(), "yt auth login") {
		t.Fatalf("Execute() error = %q, parsed query after auth", err.Error())
	}
}

func TestRawRejectsManagedHeadersBeforeAuth(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "authorization",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "raw", "/api/issues", "-H", "Authorization: Bearer bad"},
			want: "managed by credentials",
		},
		{
			name: "content type",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "raw", "/api/issues", "-H", "Content-Type: text/plain"},
			want: "ContentType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Execute(context.Background(), tt.args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil {
				t.Fatal("Execute() error = nil, want header validation")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), tt.want)
			}
			if strings.Contains(err.Error(), "yt auth login") {
				t.Fatalf("Execute() error = %q, parsed header after auth", err.Error())
			}
		})
	}
}

func TestParseRawHeadersProperties(t *testing.T) {
	property := func(value string) bool {
		headers, err := parseRawHeaders([]string{"X-Agent-Trace: " + value})
		return err == nil && headers.Get("X-Agent-Trace") == strings.TrimSpace(value)
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func TestParseRawQueryProperties(t *testing.T) {
	property := func(value string) bool {
		query, err := parseRawQuery([]string{"fields=" + value})
		return err == nil && len(query["fields"]) == 1 && query["fields"][0] == value
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func FuzzParseRawHeaders(f *testing.F) {
	for _, seed := range []string{
		"X-Agent-Trace: run-1",
		"Accept: application/xml",
		"Authorization: Bearer bad",
		"Content-Type: text/plain",
		"missing-colon",
		": empty-name",
		"X-Colon: value:with:colon",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		headers, err := parseRawHeaders([]string{raw})
		name, _, hasColon := strings.Cut(raw, ":")
		name = strings.TrimSpace(name)
		managed := strings.EqualFold(name, "Authorization") || strings.EqualFold(name, "Content-Type")
		if !hasColon || name == "" || managed || !testValidHTTPHeaderName(name) {
			if err == nil {
				t.Fatalf("parseRawHeaders(%q) succeeded for invalid or managed header %#v", raw, headers)
			}
			return
		}
		if err != nil {
			t.Fatalf("parseRawHeaders(%q) error = %v", raw, err)
		}
		if len(headers) != 1 {
			t.Fatalf("parseRawHeaders(%q) produced %d headers, want 1", raw, len(headers))
		}
	})
}

func testValidHTTPHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			continue
		}
		switch c {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}
	return true
}

func TestRawRejectsMultipleBodySources(t *testing.T) {
	err := Execute(context.Background(), []string{
		"raw", "/api/issues", "--body", `{}`, "--body-stdin",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want body source validation")
	}
	if !strings.Contains(err.Error(), "body accepts only one text source") {
		t.Fatalf("Execute() error = %q, want body source validation", err.Error())
	}
}
