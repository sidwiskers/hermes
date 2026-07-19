package api

import (
	"context"
	"fmt"
	"unicode/utf8"
)

type InputChecklistTask struct {
	ID           int             `json:"id"`
	Text         string          `json:"text"`
	ParseMode    string          `json:"parse_mode,omitempty"`
	TextEntities []MessageEntity `json:"text_entities,omitempty"`
}

type InputChecklist struct {
	Title                    string               `json:"title"`
	ParseMode                string               `json:"parse_mode,omitempty"`
	TitleEntities            []MessageEntity      `json:"title_entities,omitempty"`
	Tasks                    []InputChecklistTask `json:"tasks"`
	OthersCanAddTasks        bool                 `json:"others_can_add_tasks,omitempty"`
	OthersCanMarkTasksAsDone bool                 `json:"others_can_mark_tasks_as_done,omitempty"`
}

func validateInputChecklist(checklist InputChecklist) error {
	titleLength := utf8.RuneCountInString(checklist.Title)
	if titleLength == 0 || titleLength > 255 {
		return fmt.Errorf("hermes: checklist title must contain 1-255 characters")
	}
	if len(checklist.Tasks) == 0 || len(checklist.Tasks) > 30 {
		return fmt.Errorf("hermes: checklist requires 1-30 tasks")
	}
	ids := make(map[int]struct{}, len(checklist.Tasks))
	for _, task := range checklist.Tasks {
		length := utf8.RuneCountInString(task.Text)
		if task.ID <= 0 || length == 0 || length > 100 {
			return fmt.Errorf("hermes: checklist tasks require a positive id and 1-100 character text")
		}
		if _, exists := ids[task.ID]; exists {
			return fmt.Errorf("hermes: checklist task ids must be unique")
		}
		ids[task.ID] = struct{}{}
	}
	return nil
}

type SendChecklistParams struct {
	BusinessConnectionID string                `json:"business_connection_id"`
	ChatID               any                   `json:"chat_id"`
	Checklist            InputChecklist        `json:"checklist"`
	DisableNotification  bool                  `json:"disable_notification,omitempty"`
	ProtectContent       bool                  `json:"protect_content,omitempty"`
	MessageEffectID      string                `json:"message_effect_id,omitempty"`
	ReplyParameters      *ReplyParameters      `json:"reply_parameters,omitempty"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (client *Client) SendChecklist(ctx context.Context, params SendChecklistParams) (*Message, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "sendChecklist"); err != nil {
		return nil, err
	}
	if err := validateChatID(params.ChatID, "sendChecklist"); err != nil {
		return nil, err
	}
	if err := validateInputChecklist(params.Checklist); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "sendChecklist", params)
}

type EditMessageChecklistParams struct {
	BusinessConnectionID string                `json:"business_connection_id"`
	ChatID               any                   `json:"chat_id"`
	MessageID            int                   `json:"message_id"`
	Checklist            InputChecklist        `json:"checklist"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (client *Client) EditMessageChecklist(ctx context.Context, params EditMessageChecklistParams) (*Message, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "editMessageChecklist"); err != nil {
		return nil, err
	}
	if err := validateChatID(params.ChatID, "editMessageChecklist"); err != nil {
		return nil, err
	}
	if params.MessageID == 0 {
		return nil, fmt.Errorf("hermes: editMessageChecklist message_id is required")
	}
	if err := validateInputChecklist(params.Checklist); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "editMessageChecklist", params)
}
