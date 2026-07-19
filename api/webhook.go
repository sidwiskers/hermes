package api

import (
	"context"
	"fmt"
	"io"
)

type SetWebhookParams struct {
	URL                string   `json:"url"`
	IPAddress          string   `json:"ip_address,omitempty"`
	MaxConnections     int      `json:"max_connections,omitempty"`
	AllowedUpdates     []string `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool     `json:"drop_pending_updates,omitempty"`
	SecretToken        string   `json:"secret_token,omitempty"`
}

func (b *Client) SetWebhook(ctx context.Context, params SetWebhookParams) error {
	return b.callTrue(ctx, "setWebhook", params)
}

type DeleteWebhookParams struct {
	DropPendingUpdates bool `json:"drop_pending_updates,omitempty"`
}

func (b *Client) DeleteWebhook(ctx context.Context, params DeleteWebhookParams) error {
	return b.callTrue(ctx, "deleteWebhook", params)
}

type WebhookInfo struct {
	URL                          string   `json:"url"`
	HasCustomCertificate         bool     `json:"has_custom_certificate"`
	PendingUpdateCount           int      `json:"pending_update_count"`
	IPAddress                    string   `json:"ip_address,omitempty"`
	LastErrorDate                int64    `json:"last_error_date,omitempty"`
	LastErrorMessage             string   `json:"last_error_message,omitempty"`
	LastSynchronizationErrorDate int64    `json:"last_synchronization_error_date,omitempty"`
	MaxConnections               int      `json:"max_connections,omitempty"`
	AllowedUpdates               []string `json:"allowed_updates,omitempty"`
}

func (b *Client) GetWebhookInfo(ctx context.Context) (*WebhookInfo, error) {
	var info WebhookInfo
	if err := b.Call(ctx, "getWebhookInfo", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// SetWebhookCertificate streams a self-signed public certificate.
func (b *Client) SetWebhookCertificate(
	ctx context.Context,
	params SetWebhookParams,
	name string,
	certificate io.Reader,
) error {
	if certificate == nil {
		return fmt.Errorf("hermes: webhook certificate reader is required")
	}
	fields := formFields{}
	fields.String("url", params.URL)
	fields.String("ip_address", params.IPAddress)
	fields.Int("max_connections", params.MaxConnections)
	fields.Bool("drop_pending_updates", params.DropPendingUpdates)
	fields.String("secret_token", params.SecretToken)
	if len(params.AllowedUpdates) != 0 {
		if err := fields.JSON("allowed_updates", params.AllowedUpdates); err != nil {
			return err
		}
	}
	fields["certificate"] = "attach://certificate"
	var ok bool
	if err := b.CallMultipart(ctx, "setWebhook", fields, []Upload{{
		Field: "certificate", Name: name, Reader: certificate,
	}}, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: setWebhook returned false")
	}
	return nil
}
