package framework

import (
	"log/slog"
	"time"
)

// Logger emits one structured record after each routed update.
// Passing nil uses slog.Default().
func Logger(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next Handler) Handler {
		return func(c *Context) error {
			started := time.Now()
			err := next(c)
			attributes := []any{
				"duration", time.Since(started),
				"update_type", c.Type(),
			}
			if c != nil && c.Update != nil {
				attributes = append(attributes, "update_id", c.Update.UpdateID)
			}
			if user := c.Sender(); user != nil {
				attributes = append(attributes, "user_id", user.ID)
			}
			if chatID, ok := c.ChatID(); ok {
				attributes = append(attributes, "chat_id", chatID)
			}
			if c.Command() != "" {
				attributes = append(attributes, "command", c.Command())
			}
			if err != nil {
				attributes = append(attributes, "error", err)
				logger.Error("telegram update failed", attributes...)
			} else {
				logger.Debug("telegram update handled", attributes...)
			}
			return err
		}
	}
}
