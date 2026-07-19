package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// RichMessageMedia is media accepted by InputRichMessageMedia. The interface
// is closed to the five media types supported by Bot API 10.2.
type RichMessageMedia interface {
	richMessageMedia()
	richMessageSource() string
}

func (InputMediaAnimation) richMessageMedia()               {}
func (value InputMediaAnimation) richMessageSource() string { return value.Media }
func (InputMediaAudio) richMessageMedia()                   {}
func (value InputMediaAudio) richMessageSource() string     { return value.Media }
func (InputMediaPhoto) richMessageMedia()                   {}
func (value InputMediaPhoto) richMessageSource() string     { return value.Media }
func (InputMediaVideo) richMessageMedia()                   {}
func (value InputMediaVideo) richMessageSource() string     { return value.Media }

// InputMediaVoiceNote represents a voice note embedded in a rich message.
type InputMediaVoiceNote struct {
	Media           string          `json:"media"`
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Duration        int             `json:"duration,omitempty"`
}

func (InputMediaVoiceNote) richMessageMedia()               {}
func (value InputMediaVoiceNote) richMessageSource() string { return value.Media }
func (value InputMediaVoiceNote) MarshalJSON() ([]byte, error) {
	type plain InputMediaVoiceNote
	return marshalTaggedObject("voice_note", plain(value))
}

// InputRichMessageMedia binds an identifier used by rich HTML or Markdown to
// a concrete Telegram media value.
type InputRichMessageMedia struct {
	ID    string           `json:"id"`
	Media RichMessageMedia `json:"media"`
}

// InputRichBlock is one of the 21 block forms accepted by Bot API 10.2.
type InputRichBlock interface {
	inputRichBlock()
	richBlockType() string
}

func marshalTaggedObject(kind string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || data[0] != '{' {
		return data, nil
	}
	discriminator, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	result := make([]byte, 0, len(data)+len(discriminator)+8)
	result = append(result, `{"type":`...)
	result = append(result, discriminator...)
	if len(data) > 2 {
		result = append(result, ',')
	}
	result = append(result, data[1:]...)
	return result, nil
}

type InputRichBlockParagraph struct {
	Text RichText `json:"text"`
}

func (InputRichBlockParagraph) inputRichBlock()       {}
func (InputRichBlockParagraph) richBlockType() string { return "paragraph" }
func (value InputRichBlockParagraph) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockParagraph
	return marshalTaggedObject("paragraph", plain(value))
}

type InputRichBlockSectionHeading struct {
	Text RichText `json:"text"`
	Size int      `json:"size"`
}

func (InputRichBlockSectionHeading) inputRichBlock()       {}
func (InputRichBlockSectionHeading) richBlockType() string { return "heading" }
func (value InputRichBlockSectionHeading) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockSectionHeading
	return marshalTaggedObject("heading", plain(value))
}

type InputRichBlockPreformatted struct {
	Text     RichText `json:"text"`
	Language string   `json:"language,omitempty"`
}

func (InputRichBlockPreformatted) inputRichBlock()       {}
func (InputRichBlockPreformatted) richBlockType() string { return "pre" }
func (value InputRichBlockPreformatted) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockPreformatted
	return marshalTaggedObject("pre", plain(value))
}

type InputRichBlockFooter struct {
	Text RichText `json:"text"`
}

func (InputRichBlockFooter) inputRichBlock()       {}
func (InputRichBlockFooter) richBlockType() string { return "footer" }
func (value InputRichBlockFooter) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockFooter
	return marshalTaggedObject("footer", plain(value))
}

type InputRichBlockDivider struct{}

func (InputRichBlockDivider) inputRichBlock()       {}
func (InputRichBlockDivider) richBlockType() string { return "divider" }
func (value InputRichBlockDivider) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockDivider
	return marshalTaggedObject("divider", plain(value))
}

type InputRichBlockMathematicalExpression struct {
	Expression string `json:"expression"`
}

func (InputRichBlockMathematicalExpression) inputRichBlock() {}
func (InputRichBlockMathematicalExpression) richBlockType() string {
	return "mathematical_expression"
}
func (value InputRichBlockMathematicalExpression) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockMathematicalExpression
	return marshalTaggedObject("mathematical_expression", plain(value))
}

type InputRichBlockAnchor struct {
	Name string `json:"name"`
}

func (InputRichBlockAnchor) inputRichBlock()       {}
func (InputRichBlockAnchor) richBlockType() string { return "anchor" }
func (value InputRichBlockAnchor) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockAnchor
	return marshalTaggedObject("anchor", plain(value))
}

type InputRichBlockListItem struct {
	Blocks      []InputRichBlock `json:"blocks"`
	HasCheckbox bool             `json:"has_checkbox,omitempty"`
	IsChecked   bool             `json:"is_checked,omitempty"`
	Value       int              `json:"value,omitempty"`
	Type        string           `json:"type,omitempty"`
}

type InputRichBlockList struct {
	Items []InputRichBlockListItem `json:"items"`
}

func (InputRichBlockList) inputRichBlock()       {}
func (InputRichBlockList) richBlockType() string { return "list" }
func (value InputRichBlockList) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockList
	return marshalTaggedObject("list", plain(value))
}

type InputRichBlockBlockQuotation struct {
	Blocks []InputRichBlock `json:"blocks"`
	Credit RichText         `json:"credit,omitempty"`
}

func (InputRichBlockBlockQuotation) inputRichBlock()       {}
func (InputRichBlockBlockQuotation) richBlockType() string { return "blockquote" }
func (value InputRichBlockBlockQuotation) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockBlockQuotation
	return marshalTaggedObject("blockquote", plain(value))
}

type InputRichBlockPullQuotation struct {
	Text   RichText `json:"text"`
	Credit RichText `json:"credit,omitempty"`
}

func (InputRichBlockPullQuotation) inputRichBlock()       {}
func (InputRichBlockPullQuotation) richBlockType() string { return "pullquote" }
func (value InputRichBlockPullQuotation) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockPullQuotation
	return marshalTaggedObject("pullquote", plain(value))
}

type InputRichBlockCollage struct {
	Blocks  []InputRichBlock  `json:"blocks"`
	Caption *RichBlockCaption `json:"caption,omitempty"`
}

func (InputRichBlockCollage) inputRichBlock()       {}
func (InputRichBlockCollage) richBlockType() string { return "collage" }
func (value InputRichBlockCollage) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockCollage
	return marshalTaggedObject("collage", plain(value))
}

type InputRichBlockSlideshow struct {
	Blocks  []InputRichBlock  `json:"blocks"`
	Caption *RichBlockCaption `json:"caption,omitempty"`
}

func (InputRichBlockSlideshow) inputRichBlock()       {}
func (InputRichBlockSlideshow) richBlockType() string { return "slideshow" }
func (value InputRichBlockSlideshow) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockSlideshow
	return marshalTaggedObject("slideshow", plain(value))
}

type InputRichBlockTable struct {
	Cells      [][]RichBlockTableCell `json:"cells"`
	IsBordered bool                   `json:"is_bordered,omitempty"`
	IsStriped  bool                   `json:"is_striped,omitempty"`
	Caption    RichText               `json:"caption,omitempty"`
}

func (InputRichBlockTable) inputRichBlock()       {}
func (InputRichBlockTable) richBlockType() string { return "table" }
func (value InputRichBlockTable) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockTable
	return marshalTaggedObject("table", plain(value))
}

type InputRichBlockDetails struct {
	Summary RichText         `json:"summary"`
	Blocks  []InputRichBlock `json:"blocks"`
	IsOpen  bool             `json:"is_open,omitempty"`
}

func (InputRichBlockDetails) inputRichBlock()       {}
func (InputRichBlockDetails) richBlockType() string { return "details" }
func (value InputRichBlockDetails) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockDetails
	return marshalTaggedObject("details", plain(value))
}

type InputRichBlockMap struct {
	Location Location          `json:"location"`
	Zoom     int               `json:"zoom"`
	Width    int               `json:"width"`
	Height   int               `json:"height"`
	Caption  *RichBlockCaption `json:"caption,omitempty"`
}

func (InputRichBlockMap) inputRichBlock()       {}
func (InputRichBlockMap) richBlockType() string { return "map" }
func (value InputRichBlockMap) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockMap
	return marshalTaggedObject("map", plain(value))
}

type InputRichBlockAnimation struct {
	Animation InputMediaAnimation `json:"animation"`
	Caption   *RichBlockCaption   `json:"caption,omitempty"`
}

func (InputRichBlockAnimation) inputRichBlock()       {}
func (InputRichBlockAnimation) richBlockType() string { return "animation" }
func (value InputRichBlockAnimation) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockAnimation
	return marshalTaggedObject("animation", plain(value))
}

type InputRichBlockAudio struct {
	Audio   InputMediaAudio   `json:"audio"`
	Caption *RichBlockCaption `json:"caption,omitempty"`
}

func (InputRichBlockAudio) inputRichBlock()       {}
func (InputRichBlockAudio) richBlockType() string { return "audio" }
func (value InputRichBlockAudio) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockAudio
	return marshalTaggedObject("audio", plain(value))
}

type InputRichBlockPhoto struct {
	Photo   InputMediaPhoto   `json:"photo"`
	Caption *RichBlockCaption `json:"caption,omitempty"`
}

func (InputRichBlockPhoto) inputRichBlock()       {}
func (InputRichBlockPhoto) richBlockType() string { return "photo" }
func (value InputRichBlockPhoto) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockPhoto
	return marshalTaggedObject("photo", plain(value))
}

type InputRichBlockVideo struct {
	Video   InputMediaVideo   `json:"video"`
	Caption *RichBlockCaption `json:"caption,omitempty"`
}

func (InputRichBlockVideo) inputRichBlock()       {}
func (InputRichBlockVideo) richBlockType() string { return "video" }
func (value InputRichBlockVideo) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockVideo
	return marshalTaggedObject("video", plain(value))
}

type InputRichBlockVoiceNote struct {
	VoiceNote InputMediaVoiceNote `json:"voice_note"`
	Caption   *RichBlockCaption   `json:"caption,omitempty"`
}

func (InputRichBlockVoiceNote) inputRichBlock()       {}
func (InputRichBlockVoiceNote) richBlockType() string { return "voice_note" }
func (value InputRichBlockVoiceNote) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockVoiceNote
	return marshalTaggedObject("voice_note", plain(value))
}

type InputRichBlockThinking struct {
	Text RichText `json:"text"`
}

func (InputRichBlockThinking) inputRichBlock()       {}
func (InputRichBlockThinking) richBlockType() string { return "thinking" }
func (value InputRichBlockThinking) MarshalJSON() ([]byte, error) {
	type plain InputRichBlockThinking
	return marshalTaggedObject("thinking", plain(value))
}

type InputRichMessage struct {
	Blocks              []InputRichBlock        `json:"blocks,omitempty"`
	HTML                string                  `json:"html,omitempty"`
	Markdown            string                  `json:"markdown,omitempty"`
	Media               []InputRichMessageMedia `json:"media,omitempty"`
	IsRTL               bool                    `json:"is_rtl,omitempty"`
	SkipEntityDetection bool                    `json:"skip_entity_detection,omitempty"`
}

type InputRichMessageContent struct {
	RichMessage InputRichMessage `json:"rich_message"`
}

func validRichMediaID(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func validateRichBlocks(blocks []InputRichBlock, draft bool) error {
	for index, block := range blocks {
		if block == nil {
			return fmt.Errorf("hermes: rich block %d is nil", index)
		}
		if block.richBlockType() == "thinking" && !draft {
			return fmt.Errorf("hermes: thinking blocks are only valid in sendRichMessageDraft")
		}
		switch value := block.(type) {
		case InputRichBlockList:
			for itemIndex, item := range value.Items {
				if err := validateRichBlocks(item.Blocks, draft); err != nil {
					return fmt.Errorf("hermes: rich list item %d: %w", itemIndex, err)
				}
			}
		case InputRichBlockBlockQuotation:
			if err := validateRichBlocks(value.Blocks, draft); err != nil {
				return err
			}
		case InputRichBlockCollage:
			if err := validateRichBlocks(value.Blocks, draft); err != nil {
				return err
			}
		case InputRichBlockSlideshow:
			if err := validateRichBlocks(value.Blocks, draft); err != nil {
				return err
			}
		case InputRichBlockDetails:
			if err := validateRichBlocks(value.Blocks, draft); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRichMessage(message InputRichMessage, draft bool) error {
	formats := 0
	if len(message.Blocks) != 0 {
		formats++
	}
	if strings.TrimSpace(message.HTML) != "" {
		formats++
	}
	if strings.TrimSpace(message.Markdown) != "" {
		formats++
	}
	if formats != 1 {
		return fmt.Errorf("hermes: rich message requires exactly one of blocks, html, or markdown")
	}
	if len(message.Blocks) != 0 {
		if len(message.Media) != 0 {
			return fmt.Errorf("hermes: rich message media is only valid with html or markdown")
		}
		if err := validateRichBlocks(message.Blocks, draft); err != nil {
			return err
		}
	}
	identifiers := make(map[string]struct{}, len(message.Media))
	for index, item := range message.Media {
		if !validRichMediaID(item.ID) {
			return fmt.Errorf("hermes: rich message media %d has an invalid id", index)
		}
		if item.Media == nil || strings.TrimSpace(item.Media.richMessageSource()) == "" {
			return fmt.Errorf("hermes: rich message media %d has no media", index)
		}
		if _, exists := identifiers[item.ID]; exists {
			return fmt.Errorf("hermes: duplicate rich message media id %q", item.ID)
		}
		identifiers[item.ID] = struct{}{}
	}
	return nil
}

type SendRichMessageParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	RichMessage             InputRichMessage         `json:"rich_message"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
}

func validateSendRichMessage(params SendRichMessageParams, uploads []Upload) error {
	if err := validateChatID(params.ChatID, "sendRichMessage"); err != nil {
		return err
	}
	if err := validateRichMessage(params.RichMessage, false); err != nil {
		return err
	}
	return validateAttachmentUploads(params.RichMessage, uploads, "sendRichMessage")
}

func (client *Client) SendRichMessage(ctx context.Context, params SendRichMessageParams) (*Message, error) {
	if err := validateSendRichMessage(params, nil); err != nil {
		return nil, err
	}
	var message Message
	if err := client.Call(ctx, "sendRichMessage", params, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

// SendRichMessageUpload streams every attach:// reference nested in the rich
// message without buffering complete files in memory.
func (client *Client) SendRichMessageUpload(
	ctx context.Context,
	params SendRichMessageParams,
	uploads ...Upload,
) (*Message, error) {
	if len(uploads) == 0 {
		return client.SendRichMessage(ctx, params)
	}
	if err := validateSendRichMessage(params, uploads); err != nil {
		return nil, err
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return nil, err
	}
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_thread_id", params.MessageThreadID)
	fields.Int("direct_messages_topic_id", params.DirectMessagesTopicID)
	if err = fields.JSON("rich_message", params.RichMessage); err != nil {
		return nil, err
	}
	fields.Bool("disable_notification", params.DisableNotification)
	fields.Bool("protect_content", params.ProtectContent)
	fields.Bool("allow_paid_broadcast", params.AllowPaidBroadcast)
	fields.String("message_effect_id", params.MessageEffectID)
	if params.SuggestedPostParameters != nil {
		if err = fields.JSON("suggested_post_parameters", params.SuggestedPostParameters); err != nil {
			return nil, err
		}
	}
	if params.ReplyParameters != nil {
		if err = fields.JSON("reply_parameters", params.ReplyParameters); err != nil {
			return nil, err
		}
	}
	if params.ReplyMarkup != nil {
		if err = fields.JSON("reply_markup", params.ReplyMarkup); err != nil {
			return nil, err
		}
	}
	var message Message
	if err = client.CallMultipart(ctx, "sendRichMessage", fields, uploads, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

type SendRichMessageDraftParams struct {
	ChatID          int64            `json:"chat_id"`
	MessageThreadID int              `json:"message_thread_id,omitempty"`
	DraftID         int              `json:"draft_id"`
	RichMessage     InputRichMessage `json:"rich_message"`
}

func (client *Client) SendRichMessageDraft(ctx context.Context, params SendRichMessageDraftParams) error {
	if params.ChatID == 0 {
		return fmt.Errorf("hermes: sendRichMessageDraft requires chat_id")
	}
	if params.DraftID == 0 {
		return fmt.Errorf("hermes: sendRichMessageDraft requires a non-zero draft_id")
	}
	if err := validateRichMessage(params.RichMessage, true); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.RichMessage, nil, "sendRichMessageDraft"); err != nil {
		return err
	}
	var result bool
	return client.Call(ctx, "sendRichMessageDraft", params, &result)
}
