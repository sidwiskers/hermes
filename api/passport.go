package api

import (
	"context"
	"fmt"
	"strings"
)

type SetPassportDataErrorsParams struct {
	UserID int64                  `json:"user_id"`
	Errors []PassportElementError `json:"errors"`
}

func (client *Client) SetPassportDataErrors(ctx context.Context, params SetPassportDataErrorsParams) error {
	if params.UserID == 0 || len(params.Errors) == 0 {
		return fmt.Errorf("hermes: setPassportDataErrors requires user_id and errors")
	}
	for _, passportError := range params.Errors {
		if isNilUnion(passportError) || strings.TrimSpace(passportError.PassportElementErrorSource()) == "" ||
			strings.TrimSpace(passportError.PassportElementErrorType()) == "" ||
			strings.TrimSpace(passportError.PassportElementErrorMessage()) == "" {
			return fmt.Errorf("hermes: passport errors require source, type, and message")
		}
	}
	return client.callTrue(ctx, "setPassportDataErrors", params)
}
