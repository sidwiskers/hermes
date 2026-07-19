package api

import (
	"context"
	"testing"
)

func TestPassportElementErrorTypedNilIsRejected(t *testing.T) {
	t.Parallel()
	var passportError *PassportElementErrorDataField
	err := New("TOKEN").SetPassportDataErrors(context.Background(), SetPassportDataErrorsParams{
		UserID: 1,
		Errors: []PassportElementError{passportError},
	})
	if err == nil {
		t.Fatal("expected typed nil Passport error to fail validation")
	}
}
