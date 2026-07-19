package api

import (
	"context"
	"fmt"
)

type ManagedBotParams struct {
	UserID int64 `json:"user_id"`
}

func validateManagedBotUserID(userID int64, method string) error {
	if userID == 0 {
		return fmt.Errorf("hermes: %s user_id is required", method)
	}
	return nil
}

func (client *Client) GetManagedBotToken(ctx context.Context, params ManagedBotParams) (string, error) {
	if err := validateManagedBotUserID(params.UserID, "getManagedBotToken"); err != nil {
		return "", err
	}
	return Call[string](ctx, client, "getManagedBotToken", params)
}

func (client *Client) ReplaceManagedBotToken(ctx context.Context, params ManagedBotParams) (string, error) {
	if err := validateManagedBotUserID(params.UserID, "replaceManagedBotToken"); err != nil {
		return "", err
	}
	return Call[string](ctx, client, "replaceManagedBotToken", params)
}

func (client *Client) GetManagedBotAccessSettings(ctx context.Context, params ManagedBotParams) (BotAccessSettings, error) {
	if err := validateManagedBotUserID(params.UserID, "getManagedBotAccessSettings"); err != nil {
		return BotAccessSettings{}, err
	}
	return Call[BotAccessSettings](ctx, client, "getManagedBotAccessSettings", params)
}

type SetManagedBotAccessSettingsParams struct {
	UserID             int64   `json:"user_id"`
	IsAccessRestricted bool    `json:"is_access_restricted"`
	AddedUserIDs       []int64 `json:"added_user_ids,omitempty"`
}

func (client *Client) SetManagedBotAccessSettings(ctx context.Context, params SetManagedBotAccessSettingsParams) error {
	if err := validateManagedBotUserID(params.UserID, "setManagedBotAccessSettings"); err != nil {
		return err
	}
	if len(params.AddedUserIDs) > 10 {
		return fmt.Errorf("hermes: setManagedBotAccessSettings accepts at most 10 added users")
	}
	return client.callTrue(ctx, "setManagedBotAccessSettings", params)
}
