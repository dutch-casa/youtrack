package youtrack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: httpClient,
	}
}

type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
	Online   bool   `json:"online,omitempty"`
	Banned   bool   `json:"banned,omitempty"`
}

type Project struct {
	ID        string `json:"id,omitempty"`
	ShortName string `json:"shortName"`
	Name      string `json:"name,omitempty"`
	Archived  bool   `json:"archived,omitempty"`
	Leader    User   `json:"leader,omitempty"`
}

type Issue struct {
	ID          string          `json:"id"`
	IDReadable  string          `json:"idReadable"`
	Summary     string          `json:"summary"`
	Description string          `json:"description,omitempty"`
	Resolved    any             `json:"resolved,omitempty"`
	Project     Project         `json:"project"`
	Custom      json.RawMessage `json:"customFields,omitempty"`
}

type Comment struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Author  User   `json:"author"`
	Created int64  `json:"created"`
	Updated int64  `json:"updated,omitempty"`
}

type DurationValue struct {
	ID           string `json:"id,omitempty"`
	Minutes      int    `json:"minutes"`
	Presentation string `json:"presentation,omitempty"`
}

type WorkItemType struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type WorkItem struct {
	ID       string        `json:"id"`
	Text     string        `json:"text,omitempty"`
	Date     int64         `json:"date,omitempty"`
	Duration DurationValue `json:"duration"`
	Type     WorkItemType  `json:"type,omitempty"`
	Author   User          `json:"author,omitempty"`
	Creator  User          `json:"creator,omitempty"`
}

type Attachment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Author       User   `json:"author,omitempty"`
	Created      int64  `json:"created,omitempty"`
	Updated      int64  `json:"updated,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Extension    string `json:"extension,omitempty"`
	MimeType     string `json:"mimeType,omitempty"`
	MetaData     string `json:"metaData,omitempty"`
	URL          string `json:"url,omitempty"`
	ThumbnailURL string `json:"thumbnailURL,omitempty"`
}

type Activity struct {
	ID           string           `json:"id"`
	Type         string           `json:"$type,omitempty"`
	Author       User             `json:"author,omitempty"`
	Timestamp    int64            `json:"timestamp,omitempty"`
	Target       json.RawMessage  `json:"target,omitempty"`
	TargetMember string           `json:"targetMember,omitempty"`
	Field        ActivityField    `json:"field,omitempty"`
	Added        json.RawMessage  `json:"added,omitempty"`
	Removed      json.RawMessage  `json:"removed,omitempty"`
	Category     ActivityCategory `json:"category,omitempty"`
}

type ActivityField struct {
	Name string `json:"name,omitempty"`
}

type ActivityCategory struct {
	ID string `json:"id,omitempty"`
}

type IssueLink struct {
	ID        string   `json:"id"`
	Direction string   `json:"direction,omitempty"`
	LinkType  LinkType `json:"linkType,omitempty"`
	Issues    []Issue  `json:"issues,omitempty"`
	Trimmed   []Issue  `json:"trimmedIssues,omitempty"`
}

type LinkType struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	LocalizedName  string `json:"localizedName,omitempty"`
	SourceToTarget string `json:"sourceToTarget,omitempty"`
	TargetToSource string `json:"targetToSource,omitempty"`
	Directed       bool   `json:"directed,omitempty"`
	Aggregation    bool   `json:"aggregation,omitempty"`
}

type IssueListOptions struct {
	Query string
	Top   int
	Skip  int
}

type CreateIssueRequest struct {
	ProjectShortName string
	Summary          string
	Description      string
}

type UpdateIssueRequest struct {
	ID          string
	Summary     *string
	Description *string
}

type PageOptions struct {
	Top  int
	Skip int
}

type ApplyCommandRequest struct {
	IssueID string
	Query   string
	Comment string
	Silent  bool
}

type CommandResult struct {
	Query  string  `json:"query"`
	Issues []Issue `json:"issues"`
}

type WorkItemListOptions struct {
	IssueID string
	Top     int
	Skip    int
}

type AddWorkItemRequest struct {
	IssueID    string
	Minutes    int
	Text       string
	TypeID     string
	AuthorID   string
	DateMillis int64
	Mute       bool
}

type AttachmentListOptions struct {
	IssueID string
	Top     int
	Skip    int
}

type AttachmentFile struct {
	Name    string
	Content io.Reader
}

type UploadAttachmentsRequest struct {
	IssueID string
	Files   []AttachmentFile
}

type ActivityListOptions struct {
	IssueID    string
	Categories []string
	Top        int
	Skip       int
	Reverse    bool
	StartMs    int64
	EndMs      int64
	Author     string
}

type IssueLinkListOptions struct {
	IssueID string
	Top     int
	Skip    int
}

type RawRequest struct {
	Method      string
	Path        string
	ContentType string
	Body        io.Reader
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("youtrack api returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("youtrack api returned status %d: %s", e.StatusCode, e.Message)
}

func (c *Client) CurrentUser(ctx context.Context) (User, error) {
	var user User
	err := c.get(ctx, "/api/users/me", url.Values{
		"fields": {userFields},
	}, &user)
	return user, err
}

func (c *Client) Projects(ctx context.Context, opts PageOptions) ([]Project, error) {
	values := pageValues(opts)
	values.Set("fields", projectFields)

	var projects []Project
	err := c.get(ctx, "/api/admin/projects", values, &projects)
	return projects, err
}

func (c *Client) Users(ctx context.Context, opts PageOptions) ([]User, error) {
	values := pageValues(opts)
	values.Set("fields", userFields)

	var users []User
	err := c.get(ctx, "/api/users", values, &users)
	return users, err
}

func (c *Client) Issues(ctx context.Context, opts IssueListOptions) ([]Issue, error) {
	values := url.Values{
		"fields": {issueFields},
	}
	if opts.Query != "" {
		values.Set("query", opts.Query)
	}
	if opts.Top > 0 {
		values.Set("$top", fmt.Sprint(opts.Top))
	}
	if opts.Skip > 0 {
		values.Set("$skip", fmt.Sprint(opts.Skip))
	}

	var issues []Issue
	err := c.get(ctx, "/api/issues", values, &issues)
	return issues, err
}

func (c *Client) Issue(ctx context.Context, id string) (Issue, error) {
	if strings.TrimSpace(id) == "" {
		return Issue{}, errors.New("issue id is required")
	}
	var issue Issue
	err := c.get(ctx, "/api/issues/"+url.PathEscape(id), url.Values{
		"fields": {issueFields},
	}, &issue)
	return issue, err
}

func (c *Client) CreateIssue(ctx context.Context, req CreateIssueRequest) (Issue, error) {
	if strings.TrimSpace(req.ProjectShortName) == "" {
		return Issue{}, errors.New("project short name is required")
	}
	if strings.TrimSpace(req.Summary) == "" {
		return Issue{}, errors.New("issue summary is required")
	}
	body := map[string]any{
		"project": map[string]string{"shortName": req.ProjectShortName},
		"summary": req.Summary,
	}
	if req.Description != "" {
		body["description"] = req.Description
	}

	var issue Issue
	err := c.post(ctx, "/api/issues", url.Values{
		"fields": {issueFields},
	}, body, &issue)
	return issue, err
}

func (c *Client) UpdateIssue(ctx context.Context, req UpdateIssueRequest) (Issue, error) {
	if strings.TrimSpace(req.ID) == "" {
		return Issue{}, errors.New("issue id is required")
	}
	if req.Summary == nil && req.Description == nil {
		return Issue{}, errors.New("at least one issue field is required")
	}
	body := make(map[string]any, 2)
	if req.Summary != nil {
		body["summary"] = *req.Summary
	}
	if req.Description != nil {
		body["description"] = *req.Description
	}

	var issue Issue
	err := c.post(ctx, "/api/issues/"+url.PathEscape(req.ID), url.Values{
		"fields": {issueFields},
	}, body, &issue)
	return issue, err
}

func (c *Client) Comments(ctx context.Context, issueID string) ([]Comment, error) {
	if strings.TrimSpace(issueID) == "" {
		return nil, errors.New("issue id is required")
	}
	var comments []Comment
	err := c.get(ctx, "/api/issues/"+url.PathEscape(issueID)+"/comments", url.Values{
		"fields": {"id,text,author(id,login,name),created,updated"},
	}, &comments)
	return comments, err
}

func (c *Client) AddComment(ctx context.Context, issueID, text string) (Comment, error) {
	if strings.TrimSpace(issueID) == "" {
		return Comment{}, errors.New("issue id is required")
	}
	if strings.TrimSpace(text) == "" {
		return Comment{}, errors.New("comment text is required")
	}
	var comment Comment
	err := c.post(ctx, "/api/issues/"+url.PathEscape(issueID)+"/comments", url.Values{
		"fields": {"id,text,author(id,login,name),created,updated"},
	}, map[string]string{"text": text}, &comment)
	return comment, err
}

func (c *Client) WorkItems(ctx context.Context, opts WorkItemListOptions) ([]WorkItem, error) {
	if strings.TrimSpace(opts.IssueID) == "" {
		return nil, errors.New("issue id is required")
	}
	values := pageValues(PageOptions{Top: opts.Top, Skip: opts.Skip})
	values.Set("fields", workItemFields)

	var items []WorkItem
	err := c.get(ctx, "/api/issues/"+url.PathEscape(opts.IssueID)+"/timeTracking/workItems", values, &items)
	return items, err
}

func (c *Client) AddWorkItem(ctx context.Context, req AddWorkItemRequest) (WorkItem, error) {
	if strings.TrimSpace(req.IssueID) == "" {
		return WorkItem{}, errors.New("issue id is required")
	}
	if req.Minutes <= 0 {
		return WorkItem{}, errors.New("work item minutes must be greater than zero")
	}
	body := map[string]any{
		"duration": map[string]int{"minutes": req.Minutes},
	}
	if req.Text != "" {
		body["text"] = req.Text
	}
	if req.TypeID != "" {
		body["type"] = map[string]string{"id": req.TypeID}
	}
	if req.AuthorID != "" {
		body["author"] = map[string]string{"id": req.AuthorID}
	}
	if req.DateMillis > 0 {
		body["date"] = req.DateMillis
	}

	values := url.Values{"fields": {workItemFields}}
	if req.Mute {
		values.Set("muteUpdateNotifications", "true")
	}

	var item WorkItem
	err := c.post(ctx, "/api/issues/"+url.PathEscape(req.IssueID)+"/timeTracking/workItems", values, body, &item)
	return item, err
}

func (c *Client) Attachments(ctx context.Context, opts AttachmentListOptions) ([]Attachment, error) {
	if strings.TrimSpace(opts.IssueID) == "" {
		return nil, errors.New("issue id is required")
	}
	values := pageValues(PageOptions{Top: opts.Top, Skip: opts.Skip})
	values.Set("fields", attachmentFields)

	var attachments []Attachment
	err := c.get(ctx, "/api/issues/"+url.PathEscape(opts.IssueID)+"/attachments", values, &attachments)
	return attachments, err
}

func (c *Client) UploadAttachments(ctx context.Context, req UploadAttachmentsRequest) ([]Attachment, error) {
	if strings.TrimSpace(req.IssueID) == "" {
		return nil, errors.New("issue id is required")
	}
	if len(req.Files) == 0 {
		return nil, errors.New("at least one attachment file is required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for i, file := range req.Files {
		if strings.TrimSpace(file.Name) == "" {
			_ = writer.Close()
			return nil, fmt.Errorf("attachment file %d name is required", i+1)
		}
		if file.Content == nil {
			_ = writer.Close()
			return nil, fmt.Errorf("attachment file %q content is required", file.Name)
		}
		part, err := writer.CreateFormFile("upload", file.Name)
		if err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("create multipart part for %q: %w", file.Name, err)
		}
		if _, err := io.Copy(part, file.Content); err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("copy attachment %q: %w", file.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart body: %w", err)
	}

	values := url.Values{"fields": {attachmentFields}}
	var attachments []Attachment
	err := c.postMultipart(ctx, "/api/issues/"+url.PathEscape(req.IssueID)+"/attachments", values, writer.FormDataContentType(), &body, &attachments)
	return attachments, err
}

func (c *Client) Activities(ctx context.Context, opts ActivityListOptions) ([]Activity, error) {
	if strings.TrimSpace(opts.IssueID) == "" {
		return nil, errors.New("issue id is required")
	}
	if len(opts.Categories) == 0 {
		return nil, errors.New("at least one activity category is required")
	}
	values := pageValues(PageOptions{Top: opts.Top, Skip: opts.Skip})
	values.Set("fields", activityFields)
	values.Set("categories", strings.Join(opts.Categories, ","))
	if opts.Reverse {
		values.Set("reverse", "true")
	}
	if opts.StartMs > 0 {
		values.Set("start", fmt.Sprint(opts.StartMs))
	}
	if opts.EndMs > 0 {
		values.Set("end", fmt.Sprint(opts.EndMs))
	}
	if opts.Author != "" {
		values.Set("author", opts.Author)
	}

	var activities []Activity
	err := c.get(ctx, "/api/issues/"+url.PathEscape(opts.IssueID)+"/activities", values, &activities)
	return activities, err
}

func (c *Client) IssueLinks(ctx context.Context, opts IssueLinkListOptions) ([]IssueLink, error) {
	if strings.TrimSpace(opts.IssueID) == "" {
		return nil, errors.New("issue id is required")
	}
	values := pageValues(PageOptions{Top: opts.Top, Skip: opts.Skip})
	values.Set("fields", issueLinkFields)

	var links []IssueLink
	err := c.get(ctx, "/api/issues/"+url.PathEscape(opts.IssueID)+"/links", values, &links)
	return links, err
}

func DefaultActivityCategories() []string {
	return append([]string(nil), defaultActivityCategories...)
}

func (a Activity) Summary() string {
	parts := make([]string, 0, 4)
	if a.Field.Name != "" {
		parts = append(parts, a.Field.Name)
	}
	if a.TargetMember != "" {
		parts = append(parts, a.TargetMember)
	}
	if added := activityValues(a.Added); added != "" {
		parts = append(parts, "+"+added)
	}
	if removed := activityValues(a.Removed); removed != "" {
		parts = append(parts, "-"+removed)
	}
	if target := activityValues(a.Target); target != "" {
		parts = append(parts, target)
	}
	if len(parts) == 0 {
		return a.Type
	}
	return strings.Join(parts, " ")
}

func (c *Client) ApplyCommand(ctx context.Context, req ApplyCommandRequest) (CommandResult, error) {
	if strings.TrimSpace(req.IssueID) == "" {
		return CommandResult{}, errors.New("issue id is required")
	}
	if strings.TrimSpace(req.Query) == "" {
		return CommandResult{}, errors.New("command query is required")
	}

	body := map[string]any{
		"query":  req.Query,
		"issues": []map[string]string{{"idReadable": req.IssueID}},
		"silent": req.Silent,
	}
	if req.Comment != "" {
		body["comment"] = req.Comment
	}

	var result CommandResult
	err := c.post(ctx, "/api/commands", url.Values{
		"fields": {"query,issues(id,idReadable,summary)"},
	}, body, &result)
	return result, err
}

func (c *Client) Raw(ctx context.Context, raw RawRequest) (json.RawMessage, error) {
	if strings.TrimSpace(raw.Method) == "" {
		return nil, errors.New("raw request method is required")
	}
	if strings.TrimSpace(raw.Path) == "" {
		return nil, errors.New("raw request path is required")
	}

	req, err := c.newRequest(ctx, strings.ToUpper(raw.Method), raw.Path, nil, raw.Body)
	if err != nil {
		return nil, err
	}
	if raw.Body != nil {
		contentType := raw.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send raw request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read raw response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeAPIError(resp.StatusCode, data)
	}
	return json.RawMessage(data), nil
}

func (c *Client) get(ctx context.Context, path string, values url.Values, dst any) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, values, nil)
	if err != nil {
		return err
	}
	return c.do(req, dst)
}

func (c *Client) post(ctx context.Context, path string, values url.Values, src any, dst any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(src); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	req, err := c.newRequest(ctx, http.MethodPost, path, values, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, dst)
}

func (c *Client) postMultipart(ctx context.Context, path string, values url.Values, contentType string, body io.Reader, dst any) error {
	req, err := c.newRequest(ctx, http.MethodPost, path, values, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	return c.do(req, dst)
}

func (c *Client) newRequest(ctx context.Context, method, path string, values url.Values, body io.Reader) (*http.Request, error) {
	if c.baseURL == "" {
		return nil, errors.New("youtrack base url is empty")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	if len(values) > 0 {
		endpoint.RawQuery = values.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	return req, nil
}

func (c *Client) do(req *http.Request, dst any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp.StatusCode, data)
	}
	if dst == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func decodeAPIError(status int, data []byte) error {
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Message          string `json:"message"`
	}
	_ = json.Unmarshal(data, &payload)
	message := payload.ErrorDescription
	if message == "" {
		message = payload.Message
	}
	if message == "" {
		message = payload.Error
	}
	if message == "" {
		message = strings.TrimSpace(string(data))
	}
	return &APIError{StatusCode: status, Message: message}
}

const issueFields = "id,idReadable,summary,description,resolved,project(shortName,name),customFields(name,value(name,login,presentation,text,isResolved))"
const projectFields = "id,shortName,name,archived,leader(id,login,name,fullName,email)"
const userFields = "id,login,name,fullName,email,online,banned"
const workItemFields = "id,text,date,duration(id,minutes,presentation),type(id,name),author(id,login,name,fullName,email),creator(id,login,name,fullName,email)"
const attachmentFields = "id,name,author(id,login,name,fullName,email),created,updated,size,extension,mimeType,metaData,url,thumbnailURL"
const activityFields = "id,$type,author(id,login,name,fullName,email),timestamp,target(id,text,name,summary,idReadable),targetMember,field(name),added(id,name,login,text,presentation),removed(id,name,login,text,presentation),category(id)"
const issueLinkFields = "id,direction,linkType(id,name,localizedName,sourceToTarget,targetToSource,directed,aggregation),issues(id,idReadable,summary,resolved,project(shortName,name)),trimmedIssues(id,idReadable,summary,resolved,project(shortName,name))"

var defaultActivityCategories = []string{
	"IssueCreatedCategory",
	"SummaryCategory",
	"DescriptionCategory",
	"CustomFieldCategory",
	"CommentsCategory",
	"CommentTextCategory",
	"AttachmentsCategory",
	"LinksCategory",
	"WorkItemCategory",
	"ProjectCategory",
	"IssueResolvedCategory",
	"IssueVisibilityCategory",
	"TagsCategory",
	"VotersCategory",
	"TotalVotesCategory",
	"SprintCategory",
	"VcsChangeCategory",
	"VcsChangeStateCategory",
	"PullRequestChangeCategory",
}

func pageValues(opts PageOptions) url.Values {
	values := url.Values{}
	if opts.Top > 0 {
		values.Set("$top", fmt.Sprint(opts.Top))
	}
	if opts.Skip > 0 {
		values.Set("$skip", fmt.Sprint(opts.Skip))
	}
	return values
}

func activityValues(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}
	if value := activityValue(data); value != "" {
		return value
	}

	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err == nil {
		parts := make([]string, 0, len(values))
		for _, data := range values {
			value := activityValue(data)
			if value != "" {
				parts = append(parts, value)
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

func activityValue(data json.RawMessage) string {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return text
	}

	var object struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
		Summary      string `json:"summary"`
		IDReadable   string `json:"idReadable"`
		ID           string `json:"id"`
	}
	if err := json.Unmarshal(data, &object); err == nil {
		return firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text, object.Summary, object.IDReadable, object.ID)
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
