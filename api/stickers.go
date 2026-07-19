package api

import (
	"context"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	StickerFormatStatic   = "static"
	StickerFormatAnimated = "animated"
	StickerFormatVideo    = "video"

	StickerTypeRegular     = "regular"
	StickerTypeMask        = "mask"
	StickerTypeCustomEmoji = "custom_emoji"
)

type InputSticker struct {
	Sticker      string        `json:"sticker"`
	Format       string        `json:"format"`
	EmojiList    []string      `json:"emoji_list"`
	MaskPosition *MaskPosition `json:"mask_position,omitempty"`
	Keywords     []string      `json:"keywords,omitempty"`
}

func validStickerFormat(format string) bool {
	switch format {
	case StickerFormatStatic, StickerFormatAnimated, StickerFormatVideo:
		return true
	default:
		return false
	}
}

func validStickerType(stickerType string) bool {
	switch stickerType {
	case "", StickerTypeRegular, StickerTypeMask, StickerTypeCustomEmoji:
		return true
	default:
		return false
	}
}

func validateInputSticker(sticker InputSticker) error {
	if strings.TrimSpace(sticker.Sticker) == "" {
		return fmt.Errorf("hermes: input sticker file is required")
	}
	if !validStickerFormat(sticker.Format) {
		return fmt.Errorf("hermes: unsupported sticker format %q", sticker.Format)
	}
	if len(sticker.EmojiList) == 0 || len(sticker.EmojiList) > 20 {
		return fmt.Errorf("hermes: input sticker requires 1-20 emoji")
	}
	if len(sticker.Keywords) > 20 {
		return fmt.Errorf("hermes: input sticker accepts at most 20 keywords")
	}
	keywordLength := 0
	for _, keyword := range sticker.Keywords {
		keywordLength += utf8.RuneCountInString(keyword)
	}
	if keywordLength > 64 {
		return fmt.Errorf("hermes: input sticker keywords must not exceed 64 characters in total")
	}
	return nil
}

type GetStickerSetParams struct {
	Name string `json:"name"`
}

func (client *Client) GetStickerSet(ctx context.Context, params GetStickerSetParams) (StickerSet, error) {
	if strings.TrimSpace(params.Name) == "" {
		return StickerSet{}, fmt.Errorf("hermes: getStickerSet name is required")
	}
	return Call[StickerSet](ctx, client, "getStickerSet", params)
}

type GetCustomEmojiStickersParams struct {
	CustomEmojiIDs []string `json:"custom_emoji_ids"`
}

func (client *Client) GetCustomEmojiStickers(ctx context.Context, params GetCustomEmojiStickersParams) ([]Sticker, error) {
	if len(params.CustomEmojiIDs) == 0 || len(params.CustomEmojiIDs) > 200 {
		return nil, fmt.Errorf("hermes: getCustomEmojiStickers requires 1-200 identifiers")
	}
	return Call[[]Sticker](ctx, client, "getCustomEmojiStickers", params)
}

type UploadStickerFileParams struct {
	UserID        int64  `json:"user_id"`
	StickerFormat string `json:"sticker_format"`
}

func (client *Client) UploadStickerFile(ctx context.Context, params UploadStickerFileParams, filename string, reader io.Reader) (File, error) {
	if params.UserID == 0 {
		return File{}, fmt.Errorf("hermes: uploadStickerFile user_id is required")
	}
	if !validStickerFormat(params.StickerFormat) {
		return File{}, fmt.Errorf("hermes: unsupported sticker format %q", params.StickerFormat)
	}
	if reader == nil {
		return File{}, fmt.Errorf("hermes: uploadStickerFile reader is required")
	}
	fields := make(formFields, 2)
	fields.Int64("user_id", params.UserID)
	fields.String("sticker_format", params.StickerFormat)
	var file File
	if err := client.CallMultipart(ctx, "uploadStickerFile", fields, []Upload{{Field: "sticker", Name: filename, Reader: reader}}, &file); err != nil {
		return File{}, err
	}
	return file, nil
}

type CreateNewStickerSetParams struct {
	UserID          int64          `json:"user_id"`
	Name            string         `json:"name"`
	Title           string         `json:"title"`
	Stickers        []InputSticker `json:"stickers"`
	StickerType     string         `json:"sticker_type,omitempty"`
	NeedsRepainting bool           `json:"needs_repainting,omitempty"`
}

func validateCreateNewStickerSet(params CreateNewStickerSetParams) error {
	if params.UserID == 0 || strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("hermes: createNewStickerSet user_id and name are required")
	}
	if length := utf8.RuneCountInString(params.Title); length == 0 || length > 64 {
		return fmt.Errorf("hermes: createNewStickerSet title must contain 1-64 characters")
	}
	if len(params.Stickers) == 0 || len(params.Stickers) > 50 {
		return fmt.Errorf("hermes: createNewStickerSet requires 1-50 stickers")
	}
	if !validStickerType(params.StickerType) {
		return fmt.Errorf("hermes: unsupported sticker type %q", params.StickerType)
	}
	for _, sticker := range params.Stickers {
		if err := validateInputSticker(sticker); err != nil {
			return err
		}
	}
	return nil
}

func stickerSetFields(params CreateNewStickerSetParams) (formFields, error) {
	fields := make(formFields, 6)
	fields.Int64("user_id", params.UserID)
	fields.String("name", params.Name)
	fields.String("title", params.Title)
	if err := fields.JSON("stickers", params.Stickers); err != nil {
		return nil, err
	}
	fields.String("sticker_type", params.StickerType)
	fields.Bool("needs_repainting", params.NeedsRepainting)
	return fields, nil
}

func (client *Client) CreateNewStickerSet(ctx context.Context, params CreateNewStickerSetParams) error {
	if err := validateCreateNewStickerSet(params); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Stickers, nil, "createNewStickerSet"); err != nil {
		return err
	}
	return client.callTrue(ctx, "createNewStickerSet", params)
}

func (client *Client) CreateNewStickerSetUpload(ctx context.Context, params CreateNewStickerSetParams, uploads ...Upload) error {
	if err := validateCreateNewStickerSet(params); err != nil {
		return err
	}
	if len(uploads) == 0 {
		return client.CreateNewStickerSet(ctx, params)
	}
	if err := validateAttachmentUploads(params.Stickers, uploads, "createNewStickerSet"); err != nil {
		return err
	}
	fields, err := stickerSetFields(params)
	if err != nil {
		return err
	}
	var ok bool
	if err = client.CallMultipart(ctx, "createNewStickerSet", fields, uploads, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: createNewStickerSet returned false")
	}
	return nil
}

type AddStickerToSetParams struct {
	UserID  int64        `json:"user_id"`
	Name    string       `json:"name"`
	Sticker InputSticker `json:"sticker"`
}

func validateAddSticker(params AddStickerToSetParams, method string) error {
	if params.UserID == 0 || strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("hermes: %s user_id and name are required", method)
	}
	return validateInputSticker(params.Sticker)
}

func (client *Client) AddStickerToSet(ctx context.Context, params AddStickerToSetParams) error {
	if err := validateAddSticker(params, "addStickerToSet"); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Sticker, nil, "addStickerToSet"); err != nil {
		return err
	}
	return client.callTrue(ctx, "addStickerToSet", params)
}

func (client *Client) AddStickerToSetUpload(ctx context.Context, params AddStickerToSetParams, uploads ...Upload) error {
	if err := validateAddSticker(params, "addStickerToSet"); err != nil {
		return err
	}
	if len(uploads) == 0 {
		return client.AddStickerToSet(ctx, params)
	}
	if err := validateAttachmentUploads(params.Sticker, uploads, "addStickerToSet"); err != nil {
		return err
	}
	fields := make(formFields, 3)
	fields.Int64("user_id", params.UserID)
	fields.String("name", params.Name)
	if err := fields.JSON("sticker", params.Sticker); err != nil {
		return err
	}
	var ok bool
	if err := client.CallMultipart(ctx, "addStickerToSet", fields, uploads, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: addStickerToSet returned false")
	}
	return nil
}

type ReplaceStickerInSetParams struct {
	UserID     int64        `json:"user_id"`
	Name       string       `json:"name"`
	OldSticker string       `json:"old_sticker"`
	Sticker    InputSticker `json:"sticker"`
}

func validateReplaceSticker(params ReplaceStickerInSetParams) error {
	if params.UserID == 0 || strings.TrimSpace(params.Name) == "" || strings.TrimSpace(params.OldSticker) == "" {
		return fmt.Errorf("hermes: replaceStickerInSet user_id, name, and old_sticker are required")
	}
	return validateInputSticker(params.Sticker)
}

func (client *Client) ReplaceStickerInSet(ctx context.Context, params ReplaceStickerInSetParams) error {
	if err := validateReplaceSticker(params); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Sticker, nil, "replaceStickerInSet"); err != nil {
		return err
	}
	return client.callTrue(ctx, "replaceStickerInSet", params)
}

func (client *Client) ReplaceStickerInSetUpload(ctx context.Context, params ReplaceStickerInSetParams, uploads ...Upload) error {
	if err := validateReplaceSticker(params); err != nil {
		return err
	}
	if len(uploads) == 0 {
		return client.ReplaceStickerInSet(ctx, params)
	}
	if err := validateAttachmentUploads(params.Sticker, uploads, "replaceStickerInSet"); err != nil {
		return err
	}
	fields := make(formFields, 4)
	fields.Int64("user_id", params.UserID)
	fields.String("name", params.Name)
	fields.String("old_sticker", params.OldSticker)
	if err := fields.JSON("sticker", params.Sticker); err != nil {
		return err
	}
	var ok bool
	if err := client.CallMultipart(ctx, "replaceStickerInSet", fields, uploads, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: replaceStickerInSet returned false")
	}
	return nil
}

type StickerFileParams struct {
	Sticker string `json:"sticker"`
}

func (client *Client) DeleteStickerFromSet(ctx context.Context, params StickerFileParams) error {
	if strings.TrimSpace(params.Sticker) == "" {
		return fmt.Errorf("hermes: deleteStickerFromSet sticker is required")
	}
	return client.callTrue(ctx, "deleteStickerFromSet", params)
}

type SetStickerPositionInSetParams struct {
	Sticker  string `json:"sticker"`
	Position int    `json:"position"`
}

func (client *Client) SetStickerPositionInSet(ctx context.Context, params SetStickerPositionInSetParams) error {
	if strings.TrimSpace(params.Sticker) == "" || params.Position < 0 {
		return fmt.Errorf("hermes: setStickerPositionInSet requires sticker and a non-negative position")
	}
	return client.callTrue(ctx, "setStickerPositionInSet", params)
}

type SetStickerEmojiListParams struct {
	Sticker   string   `json:"sticker"`
	EmojiList []string `json:"emoji_list"`
}

func (client *Client) SetStickerEmojiList(ctx context.Context, params SetStickerEmojiListParams) error {
	if strings.TrimSpace(params.Sticker) == "" || len(params.EmojiList) == 0 || len(params.EmojiList) > 20 {
		return fmt.Errorf("hermes: setStickerEmojiList requires sticker and 1-20 emoji")
	}
	return client.callTrue(ctx, "setStickerEmojiList", params)
}

type SetStickerKeywordsParams struct {
	Sticker  string   `json:"sticker"`
	Keywords []string `json:"keywords,omitempty"`
}

func (client *Client) SetStickerKeywords(ctx context.Context, params SetStickerKeywordsParams) error {
	if strings.TrimSpace(params.Sticker) == "" || len(params.Keywords) > 20 {
		return fmt.Errorf("hermes: setStickerKeywords requires sticker and accepts at most 20 keywords")
	}
	total := 0
	for _, keyword := range params.Keywords {
		total += utf8.RuneCountInString(keyword)
	}
	if total > 64 {
		return fmt.Errorf("hermes: setStickerKeywords keywords must not exceed 64 characters in total")
	}
	return client.callTrue(ctx, "setStickerKeywords", params)
}

type SetStickerMaskPositionParams struct {
	Sticker      string        `json:"sticker"`
	MaskPosition *MaskPosition `json:"mask_position,omitempty"`
}

func (client *Client) SetStickerMaskPosition(ctx context.Context, params SetStickerMaskPositionParams) error {
	if strings.TrimSpace(params.Sticker) == "" {
		return fmt.Errorf("hermes: setStickerMaskPosition sticker is required")
	}
	return client.callTrue(ctx, "setStickerMaskPosition", params)
}

type SetStickerSetTitleParams struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

func (client *Client) SetStickerSetTitle(ctx context.Context, params SetStickerSetTitleParams) error {
	length := utf8.RuneCountInString(params.Title)
	if strings.TrimSpace(params.Name) == "" || length == 0 || length > 64 {
		return fmt.Errorf("hermes: setStickerSetTitle requires name and a 1-64 character title")
	}
	return client.callTrue(ctx, "setStickerSetTitle", params)
}

type SetStickerSetThumbnailParams struct {
	Name      string `json:"name"`
	UserID    int64  `json:"user_id"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Format    string `json:"format"`
}

func validateStickerSetThumbnail(params SetStickerSetThumbnailParams) error {
	if strings.TrimSpace(params.Name) == "" || params.UserID == 0 {
		return fmt.Errorf("hermes: setStickerSetThumbnail name and user_id are required")
	}
	if !validStickerFormat(params.Format) {
		return fmt.Errorf("hermes: unsupported sticker thumbnail format %q", params.Format)
	}
	return nil
}

func (client *Client) SetStickerSetThumbnail(ctx context.Context, params SetStickerSetThumbnailParams) error {
	if err := validateStickerSetThumbnail(params); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Thumbnail, nil, "setStickerSetThumbnail"); err != nil {
		return err
	}
	return client.callTrue(ctx, "setStickerSetThumbnail", params)
}

func (client *Client) SetStickerSetThumbnailUpload(ctx context.Context, params SetStickerSetThumbnailParams, upload Upload) error {
	if err := validateStickerSetThumbnail(params); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Thumbnail, []Upload{upload}, "setStickerSetThumbnail"); err != nil {
		return err
	}
	fields := make(formFields, 4)
	fields.String("name", params.Name)
	fields.Int64("user_id", params.UserID)
	fields.String("thumbnail", params.Thumbnail)
	fields.String("format", params.Format)
	var ok bool
	if err := client.CallMultipart(ctx, "setStickerSetThumbnail", fields, []Upload{upload}, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: setStickerSetThumbnail returned false")
	}
	return nil
}

type SetCustomEmojiStickerSetThumbnailParams struct {
	Name          string `json:"name"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

func (client *Client) SetCustomEmojiStickerSetThumbnail(ctx context.Context, params SetCustomEmojiStickerSetThumbnailParams) error {
	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("hermes: setCustomEmojiStickerSetThumbnail name is required")
	}
	return client.callTrue(ctx, "setCustomEmojiStickerSetThumbnail", params)
}

type DeleteStickerSetParams struct {
	Name string `json:"name"`
}

func (client *Client) DeleteStickerSet(ctx context.Context, params DeleteStickerSetParams) error {
	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("hermes: deleteStickerSet name is required")
	}
	return client.callTrue(ctx, "deleteStickerSet", params)
}
