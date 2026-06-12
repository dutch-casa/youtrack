package auth

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

const (
	EnvURL   = "YOUTRACK_URL"
	EnvToken = "YOUTRACK_TOKEN"
)

var ErrNotConfigured = errors.New("youtrack is not configured")

type Credentials struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

type Store struct {
	path string
}

func NewStore(path string) Store {
	return Store{path: path}
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, "youtrack", "config.json"), nil
}

func (s Store) Load() (Credentials, error) {
	envCreds, hasEnv := credentialsFromEnv()
	if hasEnv {
		if err := envCreds.Validate(); err != nil {
			return Credentials{}, err
		}
		return envCreds, nil
	}

	file, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Credentials{}, ErrNotConfigured
		}
		return Credentials{}, fmt.Errorf("open auth config: %w", err)
	}
	defer file.Close()

	var creds Credentials
	if err := json.NewDecoder(file).Decode(&creds); err != nil {
		return Credentials{}, fmt.Errorf("decode auth config: %w", err)
	}
	if err := creds.Validate(); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

func (s Store) Save(creds Credentials) error {
	if err := creds.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create auth config dir: %w", err)
	}
	if err := os.Chmod(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("secure auth config dir: %w", err)
	}

	tmp := s.path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create auth config: %w", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(creds); err != nil {
		file.Close()
		return fmt.Errorf("write auth config: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close auth config: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("save auth config: %w", err)
	}
	return nil
}

func (s Store) Delete() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete auth config: %w", err)
	}
	return nil
}

func Prompt(in io.Reader, out io.Writer) (Credentials, error) {
	reader := bufio.NewReader(in)
	baseURL, err := promptURL(reader, out)
	if err != nil {
		return Credentials{}, err
	}

	token, err := promptToken(in, reader, out)
	if err != nil {
		return Credentials{}, err
	}

	creds := Credentials{
		BaseURL: baseURL,
		Token:   token,
	}
	if err := creds.Validate(); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

func PromptURL(in io.Reader, out io.Writer) (string, error) {
	return promptURL(bufio.NewReader(in), out)
}

func promptURL(reader *bufio.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "YouTrack URL: ")
	baseURL, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read url: %w", err)
	}
	return strings.TrimSpace(baseURL), nil
}

func PromptToken(in io.Reader, out io.Writer) (string, error) {
	return promptToken(in, bufio.NewReader(in), out)
}

func promptToken(in io.Reader, reader *bufio.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "Permanent token: ")
	token, err := readSecret(in, reader)
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}
	fmt.Fprintln(out)
	return strings.TrimSpace(token), nil
}

func (c Credentials) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return errors.New("youtrack url is required")
	}
	parsed, err := url.Parse(c.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("youtrack url must be absolute: %q", c.BaseURL)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("youtrack url scheme must be http or https: %q", c.BaseURL)
	}
	if strings.TrimSpace(c.Token) == "" {
		return errors.New("youtrack token is required")
	}
	return nil
}

func (c Credentials) NormalizedBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
}

func credentialsFromEnv() (Credentials, bool) {
	baseURL := strings.TrimSpace(os.Getenv(EnvURL))
	token := strings.TrimSpace(os.Getenv(EnvToken))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("YT_URL"))
	}
	if token == "" {
		token = strings.TrimSpace(os.Getenv("YT_TOKEN"))
	}
	if baseURL == "" || token == "" {
		return Credentials{}, false
	}
	return Credentials{BaseURL: baseURL, Token: token}, true
}

func CanPrompt(in io.Reader) bool {
	file, ok := in.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}

func readSecret(in io.Reader, reader *bufio.Reader) (string, error) {
	file, ok := in.(*os.File)
	if ok && term.IsTerminal(int(file.Fd())) {
		secret, err := term.ReadPassword(int(file.Fd()))
		return string(secret), err
	}
	return reader.ReadString('\n')
}
