package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	StoryActive6Hours  = 6 * 60 * 60
	StoryActive12Hours = 12 * 60 * 60
	StoryActive24Hours = 24 * 60 * 60
	StoryActive48Hours = 48 * 60 * 60
)

type InputStoryContent interface {
	inputStoryContent()
	storyContentSource() string
}

type InputStoryContentPhoto struct {
	Photo string `json:"photo"`
}

func (InputStoryContentPhoto) inputStoryContent()                 {}
func (content InputStoryContentPhoto) storyContentSource() string { return content.Photo }
func (content InputStoryContentPhoto) MarshalJSON() ([]byte, error) {
	type alias InputStoryContentPhoto
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "photo", alias: alias(content)})
}

type InputStoryContentVideo struct {
	Video               string  `json:"video"`
	Duration            float64 `json:"duration,omitempty"`
	CoverFrameTimestamp float64 `json:"cover_frame_timestamp,omitempty"`
	IsAnimation         bool    `json:"is_animation,omitempty"`
}

func (InputStoryContentVideo) inputStoryContent()                 {}
func (content InputStoryContentVideo) storyContentSource() string { return content.Video }
func (content InputStoryContentVideo) MarshalJSON() ([]byte, error) {
	type alias InputStoryContentVideo
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "video", alias: alias(content)})
}

type StoryAreaPosition struct {
	XPercentage            float64 `json:"x_percentage"`
	YPercentage            float64 `json:"y_percentage"`
	WidthPercentage        float64 `json:"width_percentage"`
	HeightPercentage       float64 `json:"height_percentage"`
	RotationAngle          float64 `json:"rotation_angle"`
	CornerRadiusPercentage float64 `json:"corner_radius_percentage"`
}

// StoryAreaType is the compact union of location, suggested reaction, link,
// weather, and unique-gift story area variants.
type StoryAreaType struct {
	Type            string           `json:"type"`
	Latitude        float64          `json:"latitude,omitempty"`
	Longitude       float64          `json:"longitude,omitempty"`
	Address         *LocationAddress `json:"address,omitempty"`
	ReactionType    *ReactionType    `json:"reaction_type,omitempty"`
	IsDark          bool             `json:"is_dark,omitempty"`
	IsFlipped       bool             `json:"is_flipped,omitempty"`
	URL             string           `json:"url,omitempty"`
	Temperature     float64          `json:"temperature,omitempty"`
	Emoji           string           `json:"emoji,omitempty"`
	BackgroundColor int              `json:"background_color,omitempty"`
	Name            string           `json:"name,omitempty"`
}

type StoryArea struct {
	Position StoryAreaPosition `json:"position"`
	Type     StoryAreaType     `json:"type"`
}

type PostStoryParams struct {
	BusinessConnectionID string            `json:"business_connection_id"`
	Content              InputStoryContent `json:"content"`
	ActivePeriod         int               `json:"active_period"`
	Caption              string            `json:"caption,omitempty"`
	ParseMode            string            `json:"parse_mode,omitempty"`
	CaptionEntities      []MessageEntity   `json:"caption_entities,omitempty"`
	Areas                []StoryArea       `json:"areas,omitempty"`
	PostToChatPage       bool              `json:"post_to_chat_page,omitempty"`
	ProtectContent       bool              `json:"protect_content,omitempty"`
}

func validStoryActivePeriod(period int) bool {
	switch period {
	case StoryActive6Hours, StoryActive12Hours, StoryActive24Hours, StoryActive48Hours:
		return true
	default:
		return false
	}
}

func validateStoryContent(content InputStoryContent, method string) error {
	if content == nil || strings.TrimSpace(content.storyContentSource()) == "" {
		return fmt.Errorf("hermes: %s content is required", method)
	}
	if video, ok := content.(InputStoryContentVideo); ok && (video.Duration < 0 || video.Duration > 60) {
		return fmt.Errorf("hermes: %s video duration must be between 0 and 60 seconds", method)
	}
	return nil
}

func storyFields(params PostStoryParams) (formFields, error) {
	fields := make(formFields, 9)
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("active_period", params.ActivePeriod)
	fields.String("caption", params.Caption)
	fields.String("parse_mode", params.ParseMode)
	fields.Bool("post_to_chat_page", params.PostToChatPage)
	fields.Bool("protect_content", params.ProtectContent)
	if err := fields.JSON("content", params.Content); err != nil {
		return nil, err
	}
	if len(params.CaptionEntities) != 0 {
		if err := fields.JSON("caption_entities", params.CaptionEntities); err != nil {
			return nil, err
		}
	}
	if len(params.Areas) != 0 {
		if err := fields.JSON("areas", params.Areas); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

func (client *Client) PostStory(ctx context.Context, params PostStoryParams, uploads ...Upload) (Story, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "postStory"); err != nil {
		return Story{}, err
	}
	if err := validateStoryContent(params.Content, "postStory"); err != nil {
		return Story{}, err
	}
	if !validStoryActivePeriod(params.ActivePeriod) {
		return Story{}, fmt.Errorf("hermes: postStory active_period is unsupported")
	}
	if utf8.RuneCountInString(params.Caption) > 2048 {
		return Story{}, fmt.Errorf("hermes: postStory caption must not exceed 2048 characters")
	}
	if err := validateAttachmentUploads(params.Content, uploads, "postStory"); err != nil {
		return Story{}, err
	}
	fields, err := storyFields(params)
	if err != nil {
		return Story{}, err
	}
	var story Story
	if err = client.CallMultipart(ctx, "postStory", fields, uploads, &story); err != nil {
		return Story{}, err
	}
	return story, nil
}

type RepostStoryParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	FromChatID           int64  `json:"from_chat_id"`
	FromStoryID          int    `json:"from_story_id"`
	ActivePeriod         int    `json:"active_period"`
	PostToChatPage       bool   `json:"post_to_chat_page,omitempty"`
	ProtectContent       bool   `json:"protect_content,omitempty"`
}

func (client *Client) RepostStory(ctx context.Context, params RepostStoryParams) (Story, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "repostStory"); err != nil {
		return Story{}, err
	}
	if params.FromChatID == 0 || params.FromStoryID == 0 || !validStoryActivePeriod(params.ActivePeriod) {
		return Story{}, fmt.Errorf("hermes: repostStory requires source chat, story, and a supported active_period")
	}
	return Call[Story](ctx, client, "repostStory", params)
}

type EditStoryParams struct {
	BusinessConnectionID string            `json:"business_connection_id"`
	StoryID              int               `json:"story_id"`
	Content              InputStoryContent `json:"content"`
	Caption              string            `json:"caption,omitempty"`
	ParseMode            string            `json:"parse_mode,omitempty"`
	CaptionEntities      []MessageEntity   `json:"caption_entities,omitempty"`
	Areas                []StoryArea       `json:"areas,omitempty"`
}

func (client *Client) EditStory(ctx context.Context, params EditStoryParams, uploads ...Upload) (Story, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "editStory"); err != nil {
		return Story{}, err
	}
	if params.StoryID == 0 {
		return Story{}, fmt.Errorf("hermes: editStory story_id is required")
	}
	if err := validateStoryContent(params.Content, "editStory"); err != nil {
		return Story{}, err
	}
	if utf8.RuneCountInString(params.Caption) > 2048 {
		return Story{}, fmt.Errorf("hermes: editStory caption must not exceed 2048 characters")
	}
	if err := validateAttachmentUploads(params.Content, uploads, "editStory"); err != nil {
		return Story{}, err
	}
	fields := make(formFields, 7)
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("story_id", params.StoryID)
	fields.String("caption", params.Caption)
	fields.String("parse_mode", params.ParseMode)
	if err := fields.JSON("content", params.Content); err != nil {
		return Story{}, err
	}
	if len(params.CaptionEntities) != 0 {
		if err := fields.JSON("caption_entities", params.CaptionEntities); err != nil {
			return Story{}, err
		}
	}
	if len(params.Areas) != 0 {
		if err := fields.JSON("areas", params.Areas); err != nil {
			return Story{}, err
		}
	}
	var story Story
	if err := client.CallMultipart(ctx, "editStory", fields, uploads, &story); err != nil {
		return Story{}, err
	}
	return story, nil
}

type DeleteStoryParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	StoryID              int    `json:"story_id"`
}

func (client *Client) DeleteStory(ctx context.Context, params DeleteStoryParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "deleteStory"); err != nil {
		return err
	}
	if params.StoryID == 0 {
		return fmt.Errorf("hermes: deleteStory story_id is required")
	}
	return client.callTrue(ctx, "deleteStory", params)
}
