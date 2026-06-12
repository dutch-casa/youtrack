package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/dutchcaz/youtrack/internal/youtrack"
)

type Format string

const (
	JSON  Format = "json"
	Table Format = "table"
)

func Write(w io.Writer, format Format, value any) error {
	if format == Table {
		return writeTable(w, value)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeTable(w io.Writer, value any) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	switch rows := value.(type) {
	case []youtrack.Issue:
		fmt.Fprintln(tw, "ID\tPROJECT\tSUMMARY")
		for _, issue := range rows {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", issue.IDReadable, issue.Project.ShortName, issue.Summary)
		}
	case youtrack.Issue:
		fmt.Fprintln(tw, "FIELD\tVALUE")
		fmt.Fprintf(tw, "ID\t%s\n", rows.IDReadable)
		fmt.Fprintf(tw, "Project\t%s\n", rows.Project.ShortName)
		fmt.Fprintf(tw, "Summary\t%s\n", rows.Summary)
		fmt.Fprintf(tw, "Description\t%s\n", rows.Description)
	case []youtrack.Comment:
		fmt.Fprintln(tw, "AUTHOR\tTEXT")
		for _, comment := range rows {
			fmt.Fprintf(tw, "%s\t%s\n", comment.Author.Login, comment.Text)
		}
	case []youtrack.Project:
		fmt.Fprintln(tw, "SHORT\tNAME\tLEADER\tARCHIVED")
		for _, project := range rows {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%t\n", project.ShortName, project.Name, project.Leader.Login, project.Archived)
		}
	case []youtrack.User:
		fmt.Fprintln(tw, "LOGIN\tNAME\tEMAIL\tBANNED")
		for _, user := range rows {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%t\n", user.Login, displayName(user), user.Email, user.Banned)
		}
	case youtrack.CommandResult:
		fmt.Fprintln(tw, "QUERY\tISSUES")
		fmt.Fprintf(tw, "%s\t%d\n", rows.Query, len(rows.Issues))
	case youtrack.User:
		fmt.Fprintln(tw, "LOGIN\tNAME\tEMAIL")
		fmt.Fprintf(tw, "%s\t%s\t%s\n", rows.Login, displayName(rows), rows.Email)
	default:
		return Write(w, JSON, value)
	}
	return tw.Flush()
}

func displayName(user youtrack.User) string {
	if user.FullName != "" {
		return user.FullName
	}
	return user.Name
}
