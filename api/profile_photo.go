package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type InputProfilePhoto interface {
	inputProfilePhoto()
	profilePhotoSource() string
}

type InputProfilePhotoStatic struct {
	Photo string `json:"photo"`
}

func (InputProfilePhotoStatic) inputProfilePhoto()               {}
func (photo InputProfilePhotoStatic) profilePhotoSource() string { return photo.Photo }
func (photo InputProfilePhotoStatic) MarshalJSON() ([]byte, error) {
	type alias InputProfilePhotoStatic
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "static", alias: alias(photo)})
}

type InputProfilePhotoAnimated struct {
	Animation          string  `json:"animation"`
	MainFrameTimestamp float64 `json:"main_frame_timestamp,omitempty"`
}

func (InputProfilePhotoAnimated) inputProfilePhoto()               {}
func (photo InputProfilePhotoAnimated) profilePhotoSource() string { return photo.Animation }
func (photo InputProfilePhotoAnimated) MarshalJSON() ([]byte, error) {
	type alias InputProfilePhotoAnimated
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "animated", alias: alias(photo)})
}

type SetMyProfilePhotoParams struct {
	Photo InputProfilePhoto `json:"photo"`
}

func validateInputProfilePhoto(photo InputProfilePhoto, method string) error {
	if photo == nil || strings.TrimSpace(photo.profilePhotoSource()) == "" {
		return fmt.Errorf("hermes: %s profile photo is required", method)
	}
	return nil
}

func callMultipartTrue(ctx context.Context, client *Client, method string, fields formFields, uploads []Upload) error {
	var ok bool
	if err := client.CallMultipart(ctx, method, fields, uploads, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: %s returned false", method)
	}
	return nil
}

func (client *Client) SetMyProfilePhoto(ctx context.Context, params SetMyProfilePhotoParams, uploads ...Upload) error {
	if err := validateInputProfilePhoto(params.Photo, "setMyProfilePhoto"); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Photo, uploads, "setMyProfilePhoto"); err != nil {
		return err
	}
	fields := make(formFields, 1)
	if err := fields.JSON("photo", params.Photo); err != nil {
		return err
	}
	return callMultipartTrue(ctx, client, "setMyProfilePhoto", fields, uploads)
}

func (client *Client) RemoveMyProfilePhoto(ctx context.Context) error {
	return client.callTrue(ctx, "removeMyProfilePhoto", nil)
}
