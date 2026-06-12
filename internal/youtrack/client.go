package youtrack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (c *Client) Raw(ctx context.Context, method, path string, body io.Reader) (json.RawMessage, error) {
	req, err := c.newRequest(ctx, strings.ToUpper(method), path, nil, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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
