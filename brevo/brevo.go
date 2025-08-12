package brevo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BrevoSender represents the sender information
type BrevoSender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// BrevoRecipient represents a recipient (to, cc, or bcc)
type BrevoRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// BrevoAttachment represents an email attachment
type BrevoAttachment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// BrevoMailRequest represents the complete mail request structure
type BrevoMailRequest struct {
	Sender     BrevoSender       `json:"sender"`
	To         []BrevoRecipient  `json:"to"`
	Cc         []BrevoRecipient  `json:"cc,omitempty"`
	Bcc        []BrevoRecipient  `json:"bcc,omitempty"`
	Subject    string            `json:"subject"`
	Content    string            `json:"htmlContent,omitempty"` // Changed to match Brevo API
	Attachment []BrevoAttachment `json:"attachment,omitempty"`
	TemplateId int               `json:"templateId,omitempty"`
	Params     map[string]string `json:"params,omitempty"`
}

// parseEmailAddress parses an email address in the format "Name <email@domain.com>" or "email@domain.com"
func parseEmailAddress(address string) (name string, email string) {
	address = strings.TrimSpace(address)

	// Check if the address contains a name part
	if strings.Contains(address, "<") && strings.HasSuffix(address, ">") {
		parts := strings.Split(address, "<")
		name = strings.TrimSpace(parts[0])
		email = strings.TrimSuffix(strings.TrimSpace(parts[1]), ">")
	} else {
		email = address
		name = address // Use email as name if no name is provided
	}

	return name, email
}

// parseRecipients parses a comma-separated list of email addresses
func parseRecipients(addresses string) []BrevoRecipient {
	if addresses == "" {
		return nil
	}

	var recipients []BrevoRecipient
	addressList := strings.Split(addresses, ",")

	for _, address := range addressList {
		name, email := parseEmailAddress(strings.TrimSpace(address))
		recipients = append(recipients, BrevoRecipient{
			Email: email,
			Name:  name,
		})
	}

	return recipients
}

// readFileToBase64 reads a file and returns its base64 encoded content
func readFileToBase64(filePath string) (string, error) {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Convert to base64
	return base64.StdEncoding.EncodeToString(content), nil
}

// parseAttachments processes a newline-separated list of file paths
func parseAttachments(attachmentPaths string) ([]BrevoAttachment, error) {
	if attachmentPaths == "" {
		return nil, nil
	}

	var attachments []BrevoAttachment
	paths := strings.Split(attachmentPaths, "\n")

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		// Get the base name of the file for the attachment name
		fileName := filepath.Base(path)

		// Read and encode the file
		content, err := readFileToBase64(path)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, BrevoAttachment{
			Name:    fileName,
			Content: content,
		})
	}

	return attachments, nil
}

// getBrevoAPIKey retrieves the Brevo API key from environment variables
func getBrevoAPIKey() (string, error) {
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("BREVO_API_KEY environment variable is not set")
	}
	return apiKey, nil
}

// getBrevoAPIURL retrieves the Brevo API URL from environment variables
func getBrevoAPIURL() (string, error) {
	apiURL := os.Getenv("BREVO_API_URL")
	if apiURL == "" {
		// Provide default URL if not set
		return "https://api.brevo.com/v3/smtp/email", nil
	}
	return apiURL, nil
}

// getBrevoTemplateID retrieves the Brevo template ID from environment variables
func getBrevoTemplateID() (int, error) {
	templateIDStr := os.Getenv("BREVO_TEMPLATE_ID")
	if templateIDStr == "" {
		return 0, fmt.Errorf("BREVO_TEMPLATE_ID environment variable is not set")
	}
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		return 0, fmt.Errorf("BREVO_TEMPLATE_ID must be a valid integer: %v", err)
	}
	return templateID, nil
}

// SendApiMail sends an email using Brevo's REST API
func SendApiMail(result map[string]string) (err error) {
	// Get API key from environment
	apiKey, err := getBrevoAPIKey()
	if err != nil {
		return err
	}

	// Get API URL from environment
	apiURL, err := getBrevoAPIURL()
	if err != nil {
		return err
	}

	// Parse sender information from the 'from' field
	senderName, senderEmail := parseEmailAddress(result["from"])

	// Create the mail request
	mailReq := BrevoMailRequest{
		Sender: BrevoSender{
			Name:  senderName,
			Email: senderEmail,
		},
		Subject: result["subject"],
		Params:  make(map[string]string), // Initialize params map
	}

	// Add content if provided, splitting into MailSubLine and MailMessage
	if result["content"] != "" {
		mailReq.Params["MailMessage"] = strings.ReplaceAll(result["content"], "\n", "<br/>")

		lines := strings.SplitN(result["content"], "\n", 2)
		if len(lines) > 0 {
			// Second line goes to MailSubLine
			mailReq.Params["MailSubLine"] = strings.TrimSpace(lines[0])
		}
	}

	// Parse and add recipients
	if result["to"] != "" {
		mailReq.To = parseRecipients(result["to"])
	}

	if result["cc"] != "" {
		mailReq.Cc = parseRecipients(result["cc"])
	}

	if result["bcc"] != "" {
		mailReq.Bcc = parseRecipients(result["bcc"])
	}

	// Parse and add attachments from newline-separated file paths
	if result["attachment"] != "" {
		attachments, err := parseAttachments(result["attachment"])
		if err != nil {
			return fmt.Errorf("failed to process attachments: %v", err)
		}
		mailReq.Attachment = attachments
	}

	// Try to get template ID from result map first, then from environment if not provided
	if result["template"] != "" {
		var templateId int
		_, err = fmt.Sscanf(result["TemplateId"], "%d", &templateId)
		if err == nil {
			mailReq.TemplateId = templateId
		}
	} else {
		// If not in result map, try to get from environment
		templateId, err := getBrevoTemplateID()
		if err == nil {
			mailReq.TemplateId = templateId
		}
		// Note: We don't return error here as template ID is optional
	}
	/*
		// Add additional params if provided
		for k, v := range result {
			if k != "from" && k != "to" && k != "cc" && k != "bcc" &&
				k != "subject" && k != "content" && k != "attachment" &&
				k != "TemplateId" {
				mailReq.Params[k] = v
			}

		}*/

	// Convert request to JSON
	jsonData, err := json.Marshal(mailReq)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-key", apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
