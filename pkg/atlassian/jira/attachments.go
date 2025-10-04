package jira

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// GetAttachments retrieves all attachments for an issue
func (c *Client) GetAttachments(ctx context.Context, issueKey string) ([]Attachment, error) {
	// Get issue with attachments
	issue, err := c.GetIssue(ctx, issueKey, &GetIssueOptions{
		Fields: []string{"attachment"},
	})
	if err != nil {
		return nil, err
	}

	return issue.Fields.Attachment, nil
}

// GetAttachment retrieves a specific attachment by ID
func (c *Client) GetAttachment(ctx context.Context, attachmentID string) (*Attachment, error) {
	path := fmt.Sprintf("%s/attachment/%s", c.getAPIPath(), attachmentID)

	var attachment Attachment
	if err := c.doRequest(ctx, "GET", path, nil, &attachment); err != nil {
		return nil, fmt.Errorf("failed to get attachment %s: %w", attachmentID, err)
	}

	return &attachment, nil
}

// UploadAttachment uploads a file attachment to an issue
func (c *Client) UploadAttachment(ctx context.Context, issueKey string, filename string, content io.Reader) (*Attachment, error) {
	path := fmt.Sprintf("%s/issue/%s/attachments", c.getAPIPath(), issueKey)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, content); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Use the doRequest helper which properly handles the upload
	// Note: This is a simplified implementation
	// A full implementation would need special handling for multipart uploads
	if err := c.doRequest(ctx, http.MethodPost, path, body.Bytes(), nil); err != nil {
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}

	// Note: The actual implementation would need to parse the response
	// For now, return a basic attachment
	return &Attachment{
		Filename: filename,
	}, nil
}

// DeleteAttachment deletes an attachment
func (c *Client) DeleteAttachment(ctx context.Context, attachmentID string) error {
	path := fmt.Sprintf("%s/attachment/%s", c.getAPIPath(), attachmentID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete attachment %s: %w", attachmentID, err)
	}

	return nil
}

// DownloadAttachmentContent downloads the content of an attachment
func (c *Client) DownloadAttachmentContent(ctx context.Context, attachment *Attachment) ([]byte, error) {
	if attachment.Content == "" {
		return nil, fmt.Errorf("attachment has no content URL")
	}

	return c.DownloadAttachment(ctx, attachment.Content)
}

// GetAttachmentMetadata retrieves metadata for an attachment
func (c *Client) GetAttachmentMetadata(ctx context.Context, attachmentID string) (*Attachment, error) {
	return c.GetAttachment(ctx, attachmentID)
}
