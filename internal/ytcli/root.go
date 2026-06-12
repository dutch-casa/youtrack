package ytcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/dutchcaz/youtrack/internal/auth"
	"github.com/dutchcaz/youtrack/internal/output"
	"github.com/dutchcaz/youtrack/internal/textinput"
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
	cmd.AddCommand(a.projectsCommand())
	cmd.AddCommand(a.usersCommand())
	cmd.AddCommand(a.issuesCommand())
	cmd.AddCommand(a.commentsCommand())
	cmd.AddCommand(a.workItemsCommand())
	cmd.AddCommand(a.attachmentsCommand())
	cmd.AddCommand(a.activitiesCommand())
	cmd.AddCommand(a.linksCommand())
	cmd.AddCommand(a.commandsCommand())
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

func (a *app) projectsCommand() *cobra.Command {
	var top int
	var skip int

	list := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			projects, err := client.Projects(cmd.Context(), youtrack.PageOptions{Top: top, Skip: skip})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, projects)
		},
	}
	list.Flags().IntVar(&top, "top", 42, "maximum projects to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of projects to skip")

	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project"},
		Short:   "Work with projects",
	}
	cmd.AddCommand(list)
	return cmd
}

func (a *app) usersCommand() *cobra.Command {
	var top int
	var skip int

	list := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			users, err := client.Users(cmd.Context(), youtrack.PageOptions{Top: top, Skip: skip})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, users)
		},
	}
	list.Flags().IntVar(&top, "top", 42, "maximum users to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of users to skip")

	cmd := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Work with users",
	}
	cmd.AddCommand(list)
	return cmd
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
	var descriptionFile string
	var descriptionStdin bool
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
			resolvedDescription, _, err := textinput.Resolve(textinput.Source{
				Name:       "description",
				Literal:    description,
				LiteralSet: cmd.Flags().Changed("description"),
				File:       descriptionFile,
				Stdin:      descriptionStdin,
			}, a.in)
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			issue, err := client.CreateIssue(cmd.Context(), youtrack.CreateIssueRequest{
				ProjectShortName: project,
				Summary:          summary,
				Description:      resolvedDescription,
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
	create.Flags().StringVar(&descriptionFile, "description-file", "", "read issue description from file")
	create.Flags().BoolVar(&descriptionStdin, "description-stdin", false, "read issue description from stdin")

	var updateSummary string
	var updateDescription string
	var updateDescriptionFile string
	var updateDescriptionStdin bool
	update := &cobra.Command{
		Use:   "update ISSUE",
		Short: "Update issue summary or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			summaryChanged := cmd.Flags().Changed("summary")
			resolvedDescription, descriptionChanged, err := textinput.Resolve(textinput.Source{
				Name:       "description",
				Literal:    updateDescription,
				LiteralSet: cmd.Flags().Changed("description"),
				File:       updateDescriptionFile,
				Stdin:      updateDescriptionStdin,
			}, a.in)
			if err != nil {
				return err
			}
			if !summaryChanged && !descriptionChanged {
				return errors.New("at least one of --summary or --description is required")
			}

			var summaryValue *string
			if summaryChanged {
				summaryValue = &updateSummary
			}
			var descriptionValue *string
			if descriptionChanged {
				descriptionValue = &resolvedDescription
			}

			client, err := a.client()
			if err != nil {
				return err
			}
			issue, err := client.UpdateIssue(cmd.Context(), youtrack.UpdateIssueRequest{
				ID:          args[0],
				Summary:     summaryValue,
				Description: descriptionValue,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, issue)
		},
	}
	update.Flags().StringVarP(&updateSummary, "summary", "s", "", "new issue summary")
	update.Flags().StringVarP(&updateDescription, "description", "d", "", "new issue description")
	update.Flags().StringVar(&updateDescriptionFile, "description-file", "", "read new issue description from file")
	update.Flags().BoolVar(&updateDescriptionStdin, "description-stdin", false, "read new issue description from stdin")

	cmd := &cobra.Command{
		Use:     "issues",
		Aliases: []string{"issue"},
		Short:   "Work with issues",
	}
	cmd.AddCommand(list, show, create, update)
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
	var textFile string
	var textStdin bool
	add := &cobra.Command{
		Use:   "add ISSUE",
		Short: "Add an issue comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedText, hasText, err := textinput.Resolve(textinput.Source{
				Name:       "comment",
				Literal:    text,
				LiteralSet: cmd.Flags().Changed("text"),
				File:       textFile,
				Stdin:      textStdin,
			}, a.in)
			if err != nil {
				return err
			}
			if !hasText {
				return errors.New("--text is required")
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			comment, err := client.AddComment(cmd.Context(), args[0], resolvedText)
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, comment)
		},
	}
	add.Flags().StringVarP(&text, "text", "t", "", "comment text")
	add.Flags().StringVar(&textFile, "text-file", "", "read comment text from file")
	add.Flags().BoolVar(&textStdin, "text-stdin", false, "read comment text from stdin")

	cmd := &cobra.Command{
		Use:     "comments",
		Aliases: []string{"comment"},
		Short:   "Work with issue comments",
	}
	cmd.AddCommand(list, add)
	return cmd
}

func (a *app) workItemsCommand() *cobra.Command {
	var top int
	var skip int
	list := &cobra.Command{
		Use:   "list ISSUE",
		Short: "List issue work items",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			items, err := client.WorkItems(cmd.Context(), youtrack.WorkItemListOptions{
				IssueID: args[0],
				Top:     top,
				Skip:    skip,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, items)
		},
	}
	list.Flags().IntVar(&top, "top", 42, "maximum work items to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of work items to skip")

	var minutes int
	var text string
	var textFile string
	var textStdin bool
	var typeID string
	var authorID string
	var dateMillis int64
	var mute bool
	add := &cobra.Command{
		Use:   "add ISSUE",
		Short: "Add an issue work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if minutes <= 0 {
				return errors.New("--minutes must be greater than zero")
			}
			resolvedText, _, err := textinput.Resolve(textinput.Source{
				Name:       "work item text",
				Literal:    text,
				LiteralSet: cmd.Flags().Changed("text"),
				File:       textFile,
				Stdin:      textStdin,
			}, a.in)
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			item, err := client.AddWorkItem(cmd.Context(), youtrack.AddWorkItemRequest{
				IssueID:    args[0],
				Minutes:    minutes,
				Text:       resolvedText,
				TypeID:     typeID,
				AuthorID:   authorID,
				DateMillis: dateMillis,
				Mute:       mute,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, item)
		},
	}
	add.Flags().IntVar(&minutes, "minutes", 0, "work item duration in minutes")
	add.Flags().StringVarP(&text, "text", "t", "", "work item text")
	add.Flags().StringVar(&textFile, "text-file", "", "read work item text from file")
	add.Flags().BoolVar(&textStdin, "text-stdin", false, "read work item text from stdin")
	add.Flags().StringVar(&typeID, "type-id", "", "work item type id")
	add.Flags().StringVar(&authorID, "author-id", "", "work item author user id")
	add.Flags().Int64Var(&dateMillis, "date-ms", 0, "work item date as Unix milliseconds")
	add.Flags().BoolVar(&mute, "mute", false, "request muted update notifications")

	cmd := &cobra.Command{
		Use:     "work-items",
		Aliases: []string{"work-item", "time"},
		Short:   "Work with issue time tracking",
	}
	cmd.AddCommand(list, add)
	return cmd
}

func (a *app) attachmentsCommand() *cobra.Command {
	var top int
	var skip int
	list := &cobra.Command{
		Use:   "list ISSUE",
		Short: "List issue attachments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			attachments, err := client.Attachments(cmd.Context(), youtrack.AttachmentListOptions{
				IssueID: args[0],
				Top:     top,
				Skip:    skip,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, attachments)
		},
	}
	list.Flags().IntVar(&top, "top", 42, "maximum attachments to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of attachments to skip")

	var paths []string
	add := &cobra.Command{
		Use:   "add ISSUE",
		Short: "Attach files to an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(paths) == 0 {
				return errors.New("--file is required")
			}
			files := make([]*os.File, 0, len(paths))
			defer func() {
				for _, file := range files {
					_ = file.Close()
				}
			}()

			attachments := make([]youtrack.AttachmentFile, 0, len(paths))
			for _, path := range paths {
				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("open attachment %q: %w", path, err)
				}
				files = append(files, file)
				attachments = append(attachments, youtrack.AttachmentFile{
					Name:    filepath.Base(path),
					Content: file,
				})
			}

			client, err := a.client()
			if err != nil {
				return err
			}
			uploaded, err := client.UploadAttachments(cmd.Context(), youtrack.UploadAttachmentsRequest{
				IssueID: args[0],
				Files:   attachments,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, uploaded)
		},
	}
	add.Flags().StringArrayVarP(&paths, "file", "f", nil, "file to attach; repeat for multiple files")

	cmd := &cobra.Command{
		Use:     "attachments",
		Aliases: []string{"attachment", "files"},
		Short:   "Work with issue attachments",
	}
	cmd.AddCommand(list, add)
	return cmd
}

func (a *app) activitiesCommand() *cobra.Command {
	var top int
	var skip int
	var categories []string
	var reverse bool
	var startMs int64
	var endMs int64
	var author string

	list := &cobra.Command{
		Use:   "list ISSUE",
		Short: "List issue activity history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			effectiveCategories := categories
			if !cmd.Flags().Changed("category") {
				effectiveCategories = youtrack.DefaultActivityCategories()
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			activities, err := client.Activities(cmd.Context(), youtrack.ActivityListOptions{
				IssueID:    args[0],
				Categories: effectiveCategories,
				Top:        top,
				Skip:       skip,
				Reverse:    reverse,
				StartMs:    startMs,
				EndMs:      endMs,
				Author:     author,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, activities)
		},
	}
	list.Flags().IntVar(&top, "top", 42, "maximum activities to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of activities to skip")
	list.Flags().StringArrayVar(&categories, "category", nil, "activity category id; repeat to narrow history")
	list.Flags().BoolVar(&reverse, "reverse", true, "return newest activities first")
	list.Flags().Int64Var(&startMs, "start-ms", 0, "start timestamp as Unix milliseconds")
	list.Flags().Int64Var(&endMs, "end-ms", 0, "end timestamp as Unix milliseconds")
	list.Flags().StringVar(&author, "author", "", "filter by author id, login, Hub id, or me")

	cmd := &cobra.Command{
		Use:     "activities",
		Aliases: []string{"activity", "history"},
		Short:   "Work with issue activity history",
	}
	cmd.AddCommand(list)
	return cmd
}

func (a *app) linksCommand() *cobra.Command {
	var top int
	var skip int

	list := &cobra.Command{
		Use:   "list ISSUE",
		Short: "List issue links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			links, err := client.IssueLinks(cmd.Context(), youtrack.IssueLinkListOptions{
				IssueID: args[0],
				Top:     top,
				Skip:    skip,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, links)
		},
	}
	list.Flags().IntVar(&top, "top", 42, "maximum link buckets to return")
	list.Flags().IntVar(&skip, "skip", 0, "number of link buckets to skip")

	cmd := &cobra.Command{
		Use:     "links",
		Aliases: []string{"link"},
		Short:   "Work with issue links",
	}
	cmd.AddCommand(list)
	return cmd
}

func (a *app) commandsCommand() *cobra.Command {
	var query string
	var comment string
	var commentFile string
	var commentStdin bool
	var silent bool

	apply := &cobra.Command{
		Use:   "apply ISSUE",
		Short: "Apply a YouTrack command to an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" {
				return errors.New("--query is required")
			}
			resolvedComment, _, err := textinput.Resolve(textinput.Source{
				Name:       "comment",
				Literal:    comment,
				LiteralSet: cmd.Flags().Changed("comment"),
				File:       commentFile,
				Stdin:      commentStdin,
			}, a.in)
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.ApplyCommand(cmd.Context(), youtrack.ApplyCommandRequest{
				IssueID: args[0],
				Query:   query,
				Comment: resolvedComment,
				Silent:  silent,
			})
			if err != nil {
				return err
			}
			return output.Write(a.out, a.format, result)
		},
	}
	apply.Flags().StringVarP(&query, "query", "q", "", "YouTrack command query, for example 'State Fixed' or 'for me'")
	apply.Flags().StringVarP(&comment, "comment", "c", "", "optional command comment")
	apply.Flags().StringVar(&commentFile, "comment-file", "", "read command comment from file")
	apply.Flags().BoolVar(&commentStdin, "comment-stdin", false, "read command comment from stdin")
	apply.Flags().BoolVar(&silent, "silent", false, "apply without notifications when YouTrack permits it")

	cmd := &cobra.Command{
		Use:     "commands",
		Aliases: []string{"command", "cmd"},
		Short:   "Apply YouTrack commands",
	}
	cmd.AddCommand(apply)
	return cmd
}

func (a *app) rawCommand() *cobra.Command {
	var method string
	var contentType string
	var body string
	var bodyFile string
	var bodyStdin bool
	var headers []string
	var query []string
	var outputFile string
	cmd := &cobra.Command{
		Use:   "raw PATH",
		Short: "Call a YouTrack REST path and print the raw response",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var reader io.Reader
			resolvedBody, bodySet, err := textinput.Resolve(textinput.Source{
				Name:       "body",
				Literal:    body,
				LiteralSet: cmd.Flags().Changed("body"),
				File:       bodyFile,
				Stdin:      bodyStdin,
			}, a.in)
			if err != nil {
				return err
			}
			if bodySet {
				reader = strings.NewReader(resolvedBody)
			}
			parsedHeaders, err := parseRawHeaders(headers)
			if err != nil {
				return err
			}
			parsedQuery, err := parseRawQuery(query)
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			data, err := client.Raw(cmd.Context(), youtrack.RawRequest{
				Method:      method,
				Path:        args[0],
				Query:       parsedQuery,
				Headers:     parsedHeaders,
				ContentType: contentType,
				Body:        reader,
			})
			if err != nil {
				return err
			}
			if outputFile != "" {
				return writeRawOutputFile(outputFile, data)
			}
			_, err = fmt.Fprintln(a.out, string(data))
			return err
		},
	}
	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringVar(&contentType, "content-type", "application/json", "request body content type")
	cmd.Flags().StringVar(&body, "body", "", "request body")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read request body from file")
	cmd.Flags().BoolVar(&bodyStdin, "body-stdin", false, "read request body from stdin")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "request header as 'Name: value'; repeat for multiple headers")
	cmd.Flags().StringArrayVarP(&query, "query", "q", nil, "query parameter as name=value; repeat for multiple values")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "write raw response bytes to file")
	return cmd
}

func writeRawOutputFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func parseRawQuery(values []string) (url.Values, error) {
	query := make(url.Values)
	for _, raw := range values {
		name, value, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("raw query %q must be in name=value form", raw)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("raw query name is required")
		}
		query.Add(name, value)
	}
	return query, nil
}

func parseRawHeaders(values []string) (http.Header, error) {
	headers := make(http.Header)
	for _, raw := range values {
		name, value, ok := strings.Cut(raw, ":")
		if !ok {
			return nil, fmt.Errorf("raw header %q must be in 'Name: value' form", raw)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("raw header name is required")
		}
		if strings.EqualFold(name, "Authorization") {
			return nil, errors.New("raw header Authorization is managed by yt auth")
		}
		if strings.EqualFold(name, "Content-Type") {
			return nil, errors.New("use --content-type for Content-Type")
		}
		if !validHTTPHeaderName(name) {
			return nil, fmt.Errorf("raw header name %q is invalid", name)
		}
		headers.Add(name, strings.TrimSpace(value))
	}
	return headers, nil
}

func validHTTPHeaderName(name string) bool {
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
