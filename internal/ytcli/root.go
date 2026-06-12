package ytcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dutchcaz/youtrack/internal/auth"
	"github.com/dutchcaz/youtrack/internal/output"
	"github.com/dutchcaz/youtrack/internal/tui"
	"github.com/dutchcaz/youtrack/internal/youtrack"
	"github.com/spf13/cobra"
)

type app struct {
	in         io.Reader
	out        io.Writer
	errOut     io.Writer
	store      auth.Store
	configPath string
	format     output.Format
}

func Execute(ctx context.Context, args []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	configPath, err := auth.DefaultPath()
	if err != nil {
		return err
	}
	a := &app{
		in:         in,
		out:        out,
		errOut:     errOut,
		store:      auth.NewStore(configPath),
		configPath: configPath,
		format:     output.JSON,
	}

	cmd := a.rootCommand(ctx)
	cmd.SetArgs(args)
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	return cmd.ExecuteContext(ctx)
}

func (a *app) rootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "yt",
		Short:         "Agent-friendly YouTrack CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&a.configPath, "config", a.configPath, "auth config path")
	cmd.PersistentFlags().Var((*formatValue)(&a.format), "format", "output format: json or table")
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		a.store = auth.NewStore(a.configPath)
	}

	cmd.AddCommand(a.authCommand())
	cmd.AddCommand(a.meCommand())
	cmd.AddCommand(a.issuesCommand())
	cmd.AddCommand(a.commentsCommand())
	cmd.AddCommand(a.rawCommand())
	cmd.AddCommand(a.interactiveCommand(ctx))
	return cmd
}

func (a *app) authCommand() *cobra.Command {
	var baseURL string
	var token string

	login := &cobra.Command{
		Use:   "login",
		Short: "Save YouTrack URL and permanent token",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds := auth.Credentials{BaseURL: baseURL, Token: token}
			if creds.BaseURL == "" || creds.Token == "" {
				if !isTerminal(a.in) {
					return errors.New("non-interactive login requires --url and --token")
				}
				prompted, err := auth.Prompt(a.in, a.errOut)
				if err != nil {
					return err
				}
				if creds.BaseURL == "" {
					creds.BaseURL = prompted.BaseURL
				}
				if creds.Token == "" {
					creds.Token = prompted.Token
				}
			}
			if err := a.store.Save(creds); err != nil {
				return err
			}
			_, err := fmt.Fprintf(a.out, "saved credentials to %s\n", a.configPath)
			return err
		},
	}
	login.Flags().StringVar(&baseURL, "url", "", "YouTrack base URL, for example https://example.youtrack.cloud")
	login.Flags().StringVar(&token, "token", "", "YouTrack permanent token")

	logout := &cobra.Command{
		Use:   "logout",
		Short: "Remove saved credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.store.Delete(); err != nil {
				return err
			}
			_, err := fmt.Fprintln(a.out, "removed saved credentials")
			return err
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show whether credentials are configured",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := a.store.Load()
			if err != nil {
				if errors.Is(err, auth.ErrNotConfigured) {
					return output.Write(a.out, a.format, map[string]any{"configured": false})
				}
				return err
			}
			return output.Write(a.out, a.format, map[string]any{
				"configured": true,
				"base_url":   creds.NormalizedBaseURL(),
				"token":      redact(creds.Token),
			})
		},
	}

	cmd := &cobra.Command{Use: "auth", Short: "Manage authentication"}
	cmd.AddCommand(login, logout, status)
	return cmd
}

func (a *app) meCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Show the authenticated YouTrack user",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			user, err := client.CurrentUser(cmd.Context())
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, user)
		},
	}
}

func (a *app) issuesCommand() *cobra.Command {
	var query string
	var top int
	var skip int

	list := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			issues, err := client.Issues(cmd.Context(), youtrack.IssueListOptions{Query: query, Top: top, Skip: skip})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, issues)
		},
	}
	list.Flags().StringVarP(&query, "query", "q", "", "YouTrack issue query")
	list.Flags().IntVar(&top, "top", 25, "maximum issues to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of issues to skip")

	show := &cobra.Command{
		Use:   "show ISSUE",
		Short: "Show one issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			issue, err := client.Issue(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, issue)
		},
	}

	var project string
	var summary string
	var description string
	create := &cobra.Command{
		Use:   "create",
		Short: "Create an issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return errors.New("--project is required")
			}
			if summary == "" {
				return errors.New("--summary is required")
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			issue, err := client.CreateIssue(cmd.Context(), youtrack.CreateIssueRequest{
				ProjectShortName: project,
				Summary:          summary,
				Description:      description,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, issue)
		},
	}
	create.Flags().StringVarP(&project, "project", "p", "", "project short name")
	create.Flags().StringVarP(&summary, "summary", "s", "", "issue summary")
	create.Flags().StringVarP(&description, "description", "d", "", "issue description")

	cmd := &cobra.Command{
		Use:     "issues",
		Aliases: []string{"issue"},
		Short:   "Work with issues",
	}
	cmd.AddCommand(list, show, create)
	return cmd
}

func (a *app) commentsCommand() *cobra.Command {
	list := &cobra.Command{
		Use:   "list ISSUE",
		Short: "List issue comments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			comments, err := client.Comments(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, comments)
		},
	}

	var text string
	add := &cobra.Command{
		Use:   "add ISSUE",
		Short: "Add an issue comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if text == "" {
				return errors.New("--text is required")
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			comment, err := client.AddComment(cmd.Context(), args[0], text)
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, comment)
		},
	}
	add.Flags().StringVarP(&text, "text", "t", "", "comment text")

	cmd := &cobra.Command{
		Use:     "comments",
		Aliases: []string{"comment"},
		Short:   "Work with issue comments",
	}
	cmd.AddCommand(list, add)
	return cmd
}

func (a *app) rawCommand() *cobra.Command {
	var method string
	var body string
	cmd := &cobra.Command{
		Use:   "raw PATH",
		Short: "Call a YouTrack REST path and print raw JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			var reader io.Reader
			if body != "" {
				reader = strings.NewReader(body)
			}
			data, err := client.Raw(cmd.Context(), method, args[0], reader)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.out, string(data))
			return err
		},
	}
	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringVar(&body, "body", "", "JSON request body")
	return cmd
}

func (a *app) interactiveCommand(ctx context.Context) *cobra.Command {
	var query string
	var top int
	cmd := &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"ui", "tui"},
		Short:   "Open a lazygit-style interactive issue browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			return tui.Run(ctx, client, tui.Options{Query: query, Top: top}, a.out)
		},
	}
	cmd.Flags().StringVarP(&query, "query", "q", "", "initial YouTrack issue query")
	cmd.Flags().IntVar(&top, "top", 50, "maximum issues to load")
	return cmd
}

func (a *app) client() (*youtrack.Client, error) {
	creds, err := a.store.Ensure(a.in, a.errOut)
	if err != nil {
		return nil, err
	}
	return youtrack.NewClient(creds.NormalizedBaseURL(), creds.Token, nil), nil
}

type formatValue output.Format

func (f *formatValue) String() string {
	if f == nil || *f == "" {
		return string(output.JSON)
	}
	return string(*f)
}

func (f *formatValue) Set(value string) error {
	switch output.Format(value) {
	case output.JSON, output.Table:
		*f = formatValue(value)
		return nil
	default:
		return fmt.Errorf("unsupported format %q", value)
	}
}

func (f *formatValue) Type() string {
	return "format"
}

func isTerminal(r io.Reader) bool {
	file, ok := r.(*os.File)
	return ok && termIsTerminal(file)
}

func termIsTerminal(file *os.File) bool {
	stat, err := file.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice) != 0
}

func redact(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
