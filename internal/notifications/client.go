package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Client sends notifications via Telegram Bot API
type Client struct {
	token   string
	chatID  string
	logger  *zap.Logger
	client  *http.Client
}

// Config for Telegram client
type Config struct {
	Token  string
	ChatID string
}

// NewClient creates a new Telegram notification client
func NewClient(cfg Config, logger *zap.Logger) *Client {
	return &Client{
		token:  cfg.Token,
		chatID: cfg.ChatID,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NotificationPriority defines notification urgency (for compatibility)
type NotificationPriority string

const (
	PriorityMin     NotificationPriority = "min"
	PriorityLow     NotificationPriority = "low"
	PriorityDefault NotificationPriority = "default"
	PriorityHigh    NotificationPriority = "high"
	PriorityUrgent  NotificationPriority = "urgent"
)

// sendText sends a text message to Telegram
func (c *Client) sendText(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)

	payload := map[string]interface{}{
		"chat_id":    c.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendPhoto sends a photo with caption to Telegram
func (c *Client) sendPhoto(photoPath, caption string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", c.token)

	file, err := os.Open(photoPath)
	if err != nil {
		return fmt.Errorf("open photo: %w", err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fw, err := writer.CreateFormFile("photo", filepath.Base(photoPath))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	writer.WriteField("chat_id", c.chatID)
	if caption != "" {
		writer.WriteField("caption", caption)
		writer.WriteField("parse_mode", "Markdown")
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send photo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// formatMessage formats message with emoji and markdown
func formatMessage(emoji, title, message string, priority NotificationPriority) string {
	// Add priority indicator
	priorityEmoji := ""
	switch priority {
	case PriorityUrgent:
		priorityEmoji = "🚨 "
	case PriorityHigh:
		priorityEmoji = "⚠️ "
	}

	return fmt.Sprintf("%s%s *%s*\n%s", priorityEmoji, emoji, title, message)
}

// RunnerStarted sends notification when runner starts
func (c *Client) RunnerStarted(project, runnerID string) {
	text := formatMessage("🚀", "Runner Started",
		fmt.Sprintf("Project: `%s`\nRunner: `%s`", project, runnerID[:8]),
		PriorityDefault)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// RunnerStopped sends notification when runner stops
func (c *Client) RunnerStopped(project, runnerID string, exitCode int) {
	emoji := "✅"
	if exitCode != 0 {
		emoji = "⚠️"
	}

	text := formatMessage(emoji, "Runner Stopped",
		fmt.Sprintf("Project: `%s`\nRunner: `%s`\nExit code: `%d`", project, runnerID[:8], exitCode),
		PriorityLow)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// RunnerFailed sends notification when runner fails
func (c *Client) RunnerFailed(project, runnerID string, reason error) {
	text := formatMessage("❌", "Runner Failed",
		fmt.Sprintf("Project: `%s`\nRunner: `%s`\nReason: %v", project, runnerID[:8], reason),
		PriorityHigh)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// TokenBudgetWarning sends notification when token budget reaches threshold
func (c *Client) TokenBudgetWarning(scope string, percent int) {
	priority := PriorityDefault
	if percent >= 90 {
		priority = PriorityUrgent
	} else if percent >= 75 {
		priority = PriorityHigh
	}

	text := formatMessage("📊", "Token Budget Warning",
		fmt.Sprintf("Scope: `%s`\nUsage: *%d%%*", scope, percent),
		priority)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// DaemonStarted sends notification when daemon starts
func (c *Client) DaemonStarted(version, hostname string) {
	text := formatMessage("✨", "Stratavore Daemon Started",
		fmt.Sprintf("Version: `%s`\nHost: `%s`\nTime: %s",
			version, hostname, time.Now().Format("2006-01-02 15:04:05")),
		PriorityDefault)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// DaemonStopped sends notification when daemon stops
func (c *Client) DaemonStopped(hostname string) {
	text := formatMessage("🛑", "Stratavore Daemon Stopped",
		fmt.Sprintf("Host: `%s`\nTime: %s", hostname, time.Now().Format("2006-01-02 15:04:05")),
		PriorityDefault)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// SystemAlert sends a system-level alert
func (c *Client) SystemAlert(title, message string, priority NotificationPriority) {
	text := formatMessage("⚡", title, message, priority)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// QuotaExceeded sends notification when resource quota is exceeded
func (c *Client) QuotaExceeded(project string, resource string, limit int) {
	text := formatMessage("🚫", "Resource Quota Exceeded",
		fmt.Sprintf("Project: `%s`\nResource: `%s`\nLimit: `%d`", project, resource, limit),
		PriorityHigh)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send notification", zap.Error(err))
	}
}

// SendMetricsSummary sends a formatted metrics summary
func (c *Client) SendMetricsSummary(activeRunners, activeProjects, totalSessions int, tokensUsed, tokenLimit int64) {
	usagePercent := 0
	if tokenLimit > 0 {
		usagePercent = int((float64(tokensUsed) / float64(tokenLimit)) * 100)
	}

	text := fmt.Sprintf(`📊 *Stratavore Status Report*

🏃 Active Runners: *%d*
📁 Active Projects: *%d*
💬 Total Sessions: *%d*
🎫 Tokens Used: *%d / %d* (%d%%)

Time: %s`,
		activeRunners, activeProjects, totalSessions,
		tokensUsed, tokenLimit, usagePercent,
		time.Now().Format("2006-01-02 15:04:05"))

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send metrics summary", zap.Error(err))
	}
}

// SendCustomMessage sends a custom formatted message
func (c *Client) SendCustomMessage(emoji, title, message string) {
	text := formatMessage(emoji, title, message, PriorityDefault)

	if err := c.sendText(text); err != nil {
		c.logger.Error("failed to send custom message", zap.Error(err))
	}
}

