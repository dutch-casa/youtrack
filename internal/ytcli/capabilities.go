package ytcli

import (
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
	Completeness      []capabilityBridge      `json:"completeness"`
	Interactive       capabilityInteractive   `json:"interactive"`
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

type capabilityGuidanceRow struct {
	Need string `json:"need"`
	Use  string `json:"use"`
}

func (a *app) capabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "Describe the YouTrack CLI capability surface for agents",
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
			{Command: "yt helpdesk projects/tickets", Purpose: "Browse help desk projects and tickets.", Mutates: false},
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
		},
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
		SelectionGuidance: []capabilityGuidanceRow{
			{Need: "common stable workflow", Use: "first-class typed command"},
			{Need: "issue workflow a human would type into YouTrack command input", Use: "yt commands apply"},
			{Need: "long-tail, admin, self-hosted, or newly released REST endpoint", Use: "yt raw"},
			{Need: "human browsing", Use: "yt interactive"},
		},
	}
}
