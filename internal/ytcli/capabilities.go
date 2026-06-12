package ytcli

import (
	"strings"

	"github.com/dutch-casa/youtrack/internal/output"
	"github.com/spf13/cobra"
)

type capabilitiesDocument struct {
	SchemaVersion     int                     `json:"schemaVersion"`
	Product           string                  `json:"product"`
	CommandNames      []string                `json:"commandNames"`
	DefaultMode       string                  `json:"defaultMode"`
	DefaultOutput     string                  `json:"defaultOutput"`
	RequiresAuth      bool                    `json:"requiresAuth"`
	Authentication    []capabilityAuth        `json:"authentication"`
	FirstClass        []capabilityCommand     `json:"firstClass"`
	CommandReference  []capabilityCommandSpec `json:"commandReference"`
	Completeness      []capabilityBridge      `json:"completeness"`
	Interactive       capabilityInteractive   `json:"interactive"`
	Agent             capabilityAgent         `json:"agent"`
	SelectionGuidance []capabilityGuidanceRow `json:"selectionGuidance"`
}

type capabilityAuth struct {
	Command string `json:"command"`
	Purpose string `json:"purpose"`
}

type capabilityCommand struct {
	Command  string   `json:"command"`
	Purpose  string   `json:"purpose"`
	Mutates  bool     `json:"mutates"`
	Examples []string `json:"examples,omitempty"`
}

type capabilityCommandSpec struct {
	Command      string           `json:"command"`
	Aliases      []string         `json:"aliases,omitempty"`
	Purpose      string           `json:"purpose"`
	Args         []string         `json:"args,omitempty"`
	Flags        []capabilityFlag `json:"flags,omitempty"`
	AuthRequired bool             `json:"authRequired"`
	Mutates      bool             `json:"mutates"`
	Output       string           `json:"output"`
	Examples     []string         `json:"examples,omitempty"`
}

type capabilityFlag struct {
	Name      string `json:"name"`
	Short     string `json:"short,omitempty"`
	Value     string `json:"value,omitempty"`
	Default   string `json:"default,omitempty"`
	Required  bool   `json:"required,omitempty"`
	Repeat    bool   `json:"repeat,omitempty"`
	Purpose   string `json:"purpose"`
	Exclusive string `json:"exclusiveWith,omitempty"`
}

type capabilityBridge struct {
	Command  string   `json:"command"`
	Coverage string   `json:"coverage"`
	UseFor   []string `json:"useFor"`
}

type capabilityInteractive struct {
	Command  string   `json:"command"`
	Purpose  string   `json:"purpose"`
	Sections []string `json:"sections"`
}

type capabilityAgent struct {
	CacheKey          string   `json:"cacheKey"`
	DiscoveryCommand  string   `json:"discoveryCommand"`
	RecommendedFlow   []string `json:"recommendedFlow"`
	DoNotScrape       []string `json:"doNotScrape"`
	CompletenessRules []string `json:"completenessRules"`
}

type capabilityGuidanceRow struct {
	Need string `json:"need"`
	Use  string `json:"use"`
}

func (a *app) capabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "capabilities",
		Aliases: []string{"agent", "contract", "schema"},
		Short:   "Describe the YouTrack CLI capability surface for agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.Write(a.out, a.format, newCapabilitiesDocument())
		},
	}
}

func newCapabilitiesDocument() capabilitiesDocument {
	return capabilitiesDocument{
		SchemaVersion: 1,
		Product:       "yt",
		CommandNames:  []string{"yt", "youtrack"},
		DefaultMode:   "non-interactive",
		DefaultOutput: "json",
		RequiresAuth:  false,
		Authentication: []capabilityAuth{
			{
				Command: "yt auth login --open --url URL",
				Purpose: "Open the YouTrack Account Security page, then save a permanent token. " +
					"Use --token for non-interactive setup and YOUTRACK_URL/YOUTRACK_TOKEN for ephemeral credentials.",
			},
			{Command: "yt auth status", Purpose: "Report whether local credentials are configured without exposing token material."},
			{Command: "yt auth logout", Purpose: "Remove locally saved credentials."},
		},
		FirstClass: []capabilityCommand{
			{Command: "yt me", Purpose: "Show the authenticated user.", Mutates: false},
			{Command: "yt projects list", Purpose: "List projects.", Mutates: false},
			{Command: "yt users list", Purpose: "List users.", Mutates: false},
			{Command: "yt articles list/show", Purpose: "Browse knowledge base articles.", Mutates: false},
			{Command: "yt agiles list/sprints", Purpose: "Browse agile boards and sprints.", Mutates: false},
			{Command: "yt helpdesk projects/tickets", Purpose: "Browse help desk projects and issue-backed tickets through public project/issue resources.", Mutates: false},
			{Command: "yt issues list/show", Purpose: "Search, filter, and inspect issues.", Mutates: false},
			{Command: "yt issues create/update", Purpose: "Create issues and update stable issue fields.", Mutates: true},
			{Command: "yt comments list/add", Purpose: "Read and add issue comments.", Mutates: true},
			{Command: "yt work-items list/add", Purpose: "Read and add time tracking work items.", Mutates: true},
			{Command: "yt attachments list/add", Purpose: "Read and add issue attachments.", Mutates: true},
			{Command: "yt activities list", Purpose: "Inspect issue activity history. Alias: yt history list.", Mutates: false},
			{Command: "yt links list", Purpose: "Inspect issue links.", Mutates: false},
			{
				Command: "yt upgrade",
				Purpose: "Update the installed binary through the official install script.",
				Mutates: true,
				Examples: []string{
					"yt upgrade",
					"yt upgrade --bin-dir ~/.local/bin",
				},
			},
			{
				Command: "yt uninstall",
				Purpose: "Remove the installed binary and saved auth config.",
				Mutates: true,
				Examples: []string{
					"yt uninstall",
					"yt uninstall --bin-dir ~/.local/bin",
				},
			},
		},
		CommandReference: newCommandReference(),
		Completeness: []capabilityBridge{
			{
				Command:  "yt commands apply ISSUE --query QUERY",
				Coverage: "YouTrack command-language workflows for issues.",
				UseFor: []string{
					"state transitions",
					"assignment",
					"tags",
					"links",
					"watching/starring",
					"workflow-specific issue commands permitted by the token",
				},
			},
			{
				Command:  "yt raw PATH",
				Coverage: "Any YouTrack REST endpoint allowed by the token.",
				UseFor: []string{
					"admin endpoints",
					"new YouTrack endpoints not promoted to typed commands yet",
					"custom fields or project-specific payloads",
					"binary downloads with --output-file",
					"endpoint-specific query parameters and headers",
				},
			},
		},
		Interactive: capabilityInteractive{
			Command: "yt interactive",
			Purpose: "Optional Charm terminal workspace for human browsing and selected issue actions.",
			Sections: []string{
				"issues",
				"knowledge base",
				"help desk",
				"agile boards",
				"projects",
				"users",
			},
		},
		Agent: capabilityAgent{
			CacheKey:         "yt-capabilities-v1",
			DiscoveryCommand: "yt capabilities",
			RecommendedFlow: []string{
				"Call yt capabilities once per installed CLI version or schemaVersion and cache the JSON.",
				"Use commandReference instead of invoking --help to choose args, flags, auth requirements, mutability, output shape, and examples.",
				"Use first-class typed commands for stable common workflows.",
				"Use yt commands apply for broad issue workflows that match YouTrack's command input.",
				"Use yt raw for long-tail REST endpoints, self-hosted custom endpoints, binary downloads, or newly released YouTrack surfaces.",
			},
			DoNotScrape: []string{
				"Do not parse Cobra help output for automation.",
				"Do not infer table output; JSON is the default machine contract.",
				"Do not prompt for credentials in non-interactive runs; use yt auth login --url --token or YOUTRACK_URL/YOUTRACK_TOKEN.",
			},
			CompletenessRules: []string{
				"If a workflow has a first-class command, prefer it.",
				"If a workflow is an issue command a human would type in YouTrack, use yt commands apply.",
				"If the REST API token can do it and no first-class command exists, use yt raw.",
			},
		},
		SelectionGuidance: []capabilityGuidanceRow{
			{Need: "common stable workflow", Use: "first-class typed command"},
			{Need: "issue workflow a human would type into YouTrack command input", Use: "yt commands apply"},
			{Need: "long-tail, admin, self-hosted, or newly released REST endpoint", Use: "yt raw"},
			{Need: "human browsing", Use: "yt interactive"},
		},
	}
}

func newCommandReference() []capabilityCommandSpec {
	return []capabilityCommandSpec{
		{
			Command:      "yt capabilities",
			Purpose:      "Emit this machine-readable command contract.",
			AuthRequired: false,
			Mutates:      false,
			Output:       "JSON capability document.",
			Examples:     []string{"yt capabilities", "youtrack capabilities"},
		},
		{
			Command:      "yt auth login",
			Purpose:      "Save YouTrack credentials.",
			Flags:        []capabilityFlag{requiredFlag("url", "", "URL", "YouTrack base URL for non-interactive setup"), requiredFlag("token", "", "TOKEN", "Permanent token for non-interactive setup"), boolFlag("open", "", "Open the account security page before token entry"), boolFlag("no-verify", "", "Save without calling YouTrack to verify the token")},
			AuthRequired: false,
			Mutates:      true,
			Output:       "JSON object with saved/configPath/verified and optional user.",
			Examples:     []string{"yt auth login --open --url https://example.youtrack.cloud", "yt auth login --url https://example.youtrack.cloud --token perm:..."},
		},
		{
			Command:      "yt auth status",
			Purpose:      "Report local credential configuration with token redaction.",
			AuthRequired: false,
			Mutates:      false,
			Output:       "JSON object with configured status and redacted token when present.",
			Examples:     []string{"yt auth status"},
		},
		{
			Command:      "yt auth logout",
			Purpose:      "Remove saved credentials.",
			AuthRequired: false,
			Mutates:      true,
			Output:       "JSON object with removed true.",
			Examples:     []string{"yt auth logout"},
		},
		{
			Command:      "yt me",
			Purpose:      "Show the authenticated user.",
			AuthRequired: true,
			Mutates:      false,
			Output:       "YouTrack user JSON.",
			Examples:     []string{"yt me"},
		},
		listSpec("yt projects list", nil, "List projects.", []string{"yt projects list --top 100"}),
		listSpec("yt users list", nil, "List users.", []string{"yt users list --top 100"}),
		listSpec("yt articles list", []capabilityFlag{valueFlag("project", "p", "PROJECT", "", "Restrict articles to a project short name or id")}, "List knowledge base articles.", []string{"yt articles list --project ABC"}),
		{
			Command:      "yt articles show ARTICLE",
			Aliases:      []string{"yt kb show ARTICLE", "yt knowledge-base show ARTICLE"},
			Purpose:      "Show one knowledge base article.",
			Args:         []string{"ARTICLE"},
			AuthRequired: true,
			Mutates:      false,
			Output:       "YouTrack article JSON.",
			Examples:     []string{"yt articles show ABC-A-1"},
		},
		listSpec("yt agiles list", nil, "List agile boards.", []string{"yt agiles list"}),
		listSpec("yt agiles sprints AGILE", nil, "List sprints for an agile board.", []string{"yt agiles sprints 120-1"}),
		listSpec("yt helpdesk projects", nil, "List help desk projects.", []string{"yt helpdesk projects"}),
		listSpec("yt helpdesk tickets PROJECT", []capabilityFlag{valueFlag("query", "q", "QUERY", "", "Additional YouTrack issue query")}, "List issue-backed help desk tickets in a project. Use yt raw for Helpdesk-specific endpoints or app-provided custom endpoints.", []string{"yt helpdesk tickets SUPPORT --query '#Unresolved'"}),
		listSpec("yt issues list", []capabilityFlag{valueFlag("query", "q", "QUERY", "", "YouTrack issue query")}, "Search and list issues.", []string{"yt issues list --query 'project: ABC #Unresolved' --top 20"}),
		{
			Command:      "yt issues show ISSUE",
			Purpose:      "Show one issue.",
			Args:         []string{"ISSUE"},
			AuthRequired: true,
			Mutates:      false,
			Output:       "YouTrack issue JSON.",
			Examples:     []string{"yt issues show ABC-123"},
		},
		{
			Command: "yt issues create",
			Purpose: "Create an issue.",
			Flags: []capabilityFlag{
				requiredFlag("project", "p", "PROJECT", "Project short name"),
				requiredFlag("summary", "s", "TEXT", "Issue summary"),
				textFlag("description", "d", "TEXT", "Issue description"),
				textSourceFlag("description-file", "PATH", "Read issue description from file", "description,description-stdin"),
				boolTextSourceFlag("description-stdin", "Read issue description from stdin", "description,description-file"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Created YouTrack issue JSON.",
			Examples:     []string{"yt issues create --project ABC --summary 'Fix login redirect'", "yt issues create --project ABC --summary 'Long report' --description-file ./report.md"},
		},
		{
			Command: "yt issues update ISSUE",
			Purpose: "Update issue summary or description.",
			Args:    []string{"ISSUE"},
			Flags: []capabilityFlag{
				textFlag("summary", "s", "TEXT", "New issue summary"),
				textFlag("description", "d", "TEXT", "New issue description"),
				textSourceFlag("description-file", "PATH", "Read new issue description from file", "description,description-stdin"),
				boolTextSourceFlag("description-stdin", "Read new issue description from stdin", "description,description-file"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Updated YouTrack issue JSON.",
			Examples:     []string{"yt issues update ABC-123 --summary 'Fix login redirect after SSO'"},
		},
		{
			Command:      "yt comments list ISSUE",
			Purpose:      "List issue comments.",
			Args:         []string{"ISSUE"},
			AuthRequired: true,
			Mutates:      false,
			Output:       "Array of YouTrack comment JSON objects.",
			Examples:     []string{"yt comments list ABC-123"},
		},
		{
			Command: "yt comments add ISSUE",
			Purpose: "Add an issue comment.",
			Args:    []string{"ISSUE"},
			Flags: []capabilityFlag{
				requiredFlag("text", "t", "TEXT", "Comment text"),
				textSourceFlag("text-file", "PATH", "Read comment text from file", "text,text-stdin"),
				boolTextSourceFlag("text-stdin", "Read comment text from stdin", "text,text-file"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Created YouTrack comment JSON.",
			Examples:     []string{"yt comments add ABC-123 --text 'I can reproduce this.'", "yt comments add ABC-123 --text-stdin < ./notes.md"},
		},
		listSpec("yt work-items list ISSUE", nil, "List issue work items.", []string{"yt work-items list ABC-123"}),
		{
			Command: "yt work-items add ISSUE",
			Purpose: "Add an issue work item.",
			Args:    []string{"ISSUE"},
			Flags: []capabilityFlag{
				requiredFlag("minutes", "", "MINUTES", "Work item duration in minutes; must be greater than zero"),
				textFlag("text", "t", "TEXT", "Work item text"),
				textSourceFlag("text-file", "PATH", "Read work item text from file", "text,text-stdin"),
				boolTextSourceFlag("text-stdin", "Read work item text from stdin", "text,text-file"),
				valueFlag("type-id", "", "ID", "", "Work item type id"),
				valueFlag("author-id", "", "ID", "", "Work item author user id"),
				valueFlag("date-ms", "", "MILLIS", "0", "Work item date as Unix milliseconds"),
				boolFlag("mute", "", "Request muted update notifications"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Created YouTrack work item JSON.",
			Examples:     []string{"yt work-items add ABC-123 --minutes 45 --text 'implementation'"},
		},
		listSpec("yt attachments list ISSUE", nil, "List issue attachments.", []string{"yt attachments list ABC-123"}),
		{
			Command:      "yt attachments add ISSUE",
			Purpose:      "Attach one or more files to an issue.",
			Args:         []string{"ISSUE"},
			Flags:        []capabilityFlag{requiredRepeatFlag("file", "f", "PATH", "File to attach; repeat for multiple files")},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Array of uploaded YouTrack attachment JSON objects.",
			Examples:     []string{"yt attachments add ABC-123 --file ./screenshot.png"},
		},
		listSpec("yt activities list ISSUE", []capabilityFlag{repeatFlag("category", "", "CATEGORY", "Activity category id; repeat to narrow history"), boolDefaultFlag("reverse", "", "true", "Return newest activities first"), valueFlag("start-ms", "", "MILLIS", "0", "Start timestamp as Unix milliseconds"), valueFlag("end-ms", "", "MILLIS", "0", "End timestamp as Unix milliseconds"), valueFlag("author", "", "USER", "", "Filter by author id, login, Hub id, or me")}, "List issue activity history. Alias: yt history list.", []string{"yt history list ABC-123 --category CommentsCategory --category CustomFieldCategory"}),
		listSpec("yt links list ISSUE", nil, "List issue links.", []string{"yt links list ABC-123"}),
		{
			Command: "yt commands apply ISSUE",
			Purpose: "Apply a YouTrack command-language query to an issue.",
			Args:    []string{"ISSUE"},
			Flags: []capabilityFlag{
				requiredFlag("query", "q", "QUERY", "YouTrack command query"),
				textFlag("comment", "c", "TEXT", "Optional command comment"),
				textSourceFlag("comment-file", "PATH", "Read command comment from file", "comment,comment-stdin"),
				boolTextSourceFlag("comment-stdin", "Read command comment from stdin", "comment,comment-file"),
				boolFlag("silent", "", "Apply without notifications when YouTrack permits it"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "YouTrack command result JSON.",
			Examples:     []string{"yt commands apply ABC-123 --query 'State Fixed' --comment 'Fixed in main'"},
		},
		{
			Command: "yt raw PATH",
			Purpose: "Call any YouTrack REST path and write exact response bytes.",
			Args:    []string{"PATH"},
			Flags: []capabilityFlag{
				valueFlag("method", "X", "METHOD", "GET", "HTTP method"),
				valueFlag("content-type", "", "TYPE", "application/json", "Request body content type"),
				textFlag("body", "", "TEXT", "Request body"),
				textSourceFlag("body-file", "PATH", "Read request body from file", "body,body-stdin"),
				boolTextSourceFlag("body-stdin", "Read request body from stdin", "body,body-file"),
				repeatFlag("header", "H", "NAME: VALUE", "Endpoint-specific request header; Authorization and Content-Type are managed"),
				repeatFlag("query", "q", "NAME=VALUE", "Query parameter; repeat for multiple values"),
				valueFlag("output-file", "o", "PATH", "", "Write raw response bytes to file"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Exact response bytes to stdout or --output-file.",
			Examples:     []string{"yt raw /api/admin/projects", "yt raw /api/issues --method POST --body-file ./issue.json", "yt raw /api/files/123 --output-file ./download.bin"},
		},
		{
			Command: "yt interactive",
			Aliases: []string{
				"yt ui",
				"yt tui",
			},
			Purpose: "Open the optional Charm terminal workspace.",
			Flags: []capabilityFlag{
				valueFlag("query", "q", "QUERY", "", "Initial YouTrack issue query"),
				valueFlag("top", "", "N", "50", "Maximum issues to load"),
				valueFlag("image-protocol", "", "auto|none|kitty|iterm2", "auto", "Terminal image protocol for attachment previews"),
			},
			AuthRequired: true,
			Mutates:      true,
			Output:       "Interactive terminal UI.",
			Examples:     []string{"yt interactive --query 'project: ABC #Unresolved'"},
		},
		{
			Command: "yt upgrade",
			Purpose: "Update the installed binary with the public installer.",
			Flags: []capabilityFlag{
				valueFlag("bin-dir", "", "DIR", "", "Directory to install into; defaults to this executable's directory"),
				valueFlag("name", "", "NAME", "", "Installed binary name; defaults to this executable's file name"),
				valueFlag("installer-url", "", "URL", defaultInstallerURL, "Installer script URL"),
			},
			AuthRequired: false,
			Mutates:      true,
			Output:       "JSON object with updated true and path.",
			Examples:     []string{"yt upgrade", "youtrack upgrade"},
		},
		{
			Command: "yt uninstall",
			Purpose: "Remove the installed binary, its paired alias when present, and the saved auth config.",
			Flags: []capabilityFlag{
				valueFlag("bin-dir", "", "DIR", "", "Directory containing the installed binary; defaults to this executable's directory"),
				valueFlag("name", "", "NAME", "", "Installed binary name; defaults to this executable's file name"),
			},
			AuthRequired: false,
			Mutates:      true,
			Output:       "JSON object with removed true, configRemoved true, path, configPath, and optional aliasPath.",
			Examples:     []string{"yt uninstall", "youtrack uninstall"},
		},
	}
}

func listSpec(command string, extraFlags []capabilityFlag, purpose string, examples []string) capabilityCommandSpec {
	flags := []capabilityFlag{
		valueFlag("top", "", "N", listDefaultFor(command), "Maximum rows to return"),
		valueFlag("skip", "", "N", "0", "Number of rows to skip"),
	}
	flags = append(extraFlags, flags...)
	return capabilityCommandSpec{
		Command:      command,
		Purpose:      purpose,
		Args:         commandArgs(command),
		Flags:        flags,
		AuthRequired: true,
		Mutates:      false,
		Output:       "JSON array.",
		Examples:     examples,
	}
}

func commandArgs(command string) []string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return nil
	}
	last := fields[len(fields)-1]
	if last == strings.ToUpper(last) && strings.ContainsAny(last, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return []string{last}
	}
	return nil
}

func listDefaultFor(command string) string {
	if strings.Contains(command, "issues list") || strings.Contains(command, "helpdesk tickets") {
		return "25"
	}
	return "42"
}

func requiredFlag(name, short, value, purpose string) capabilityFlag {
	flag := valueFlag(name, short, value, "", purpose)
	flag.Required = true
	return flag
}

func requiredRepeatFlag(name, short, value, purpose string) capabilityFlag {
	flag := requiredFlag(name, short, value, purpose)
	flag.Repeat = true
	return flag
}

func repeatFlag(name, short, value, purpose string) capabilityFlag {
	flag := valueFlag(name, short, value, "", purpose)
	flag.Repeat = true
	return flag
}

func textFlag(name, short, value, purpose string) capabilityFlag {
	return valueFlag(name, short, value, "", purpose)
}

func textSourceFlag(name, value, purpose, exclusive string) capabilityFlag {
	flag := valueFlag(name, "", value, "", purpose)
	flag.Exclusive = exclusive
	return flag
}

func boolTextSourceFlag(name, purpose, exclusive string) capabilityFlag {
	flag := boolFlag(name, "", purpose)
	flag.Exclusive = exclusive
	return flag
}

func boolDefaultFlag(name, short, defaultValue, purpose string) capabilityFlag {
	flag := boolFlag(name, short, purpose)
	flag.Default = defaultValue
	return flag
}

func boolFlag(name, short, purpose string) capabilityFlag {
	return capabilityFlag{Name: name, Short: short, Value: "bool", Purpose: purpose}
}

func valueFlag(name, short, value, defaultValue, purpose string) capabilityFlag {
	return capabilityFlag{Name: name, Short: short, Value: value, Default: defaultValue, Purpose: purpose}
}
