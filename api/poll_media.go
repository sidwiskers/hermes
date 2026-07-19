package api

import (
	"encoding/json"
	"strings"
)

// InputPollMedia is media accepted in a poll description or quiz explanation.
// The interface is closed so only Bot API-compatible media values can be used.
type InputPollMedia interface {
	inputPollMedia()
	validPollMedia() bool
}

// InputPollOptionMedia is media accepted on a poll option.
type InputPollOptionMedia interface {
	inputPollOptionMedia()
	validPollMedia() bool
}

// InputMediaAnimation represents an animation in poll media.
type InputMediaAnimation struct {
	Media                 string          `json:"media"`
	Thumbnail             string          `json:"thumbnail,omitempty"`
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	Width                 int             `json:"width,omitempty"`
	Height                int             `json:"height,omitempty"`
	Duration              int             `json:"duration,omitempty"`
	HasSpoiler            bool            `json:"has_spoiler,omitempty"`
}

func (InputMediaAnimation) inputPollMedia()       {}
func (InputMediaAnimation) inputPollOptionMedia() {}
func (m InputMediaAnimation) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != ""
}
func (m InputMediaAnimation) MarshalJSON() ([]byte, error) {
	type alias InputMediaAnimation
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "animation", alias: alias(m)})
}

// InputMediaLivePhoto represents a live photo. Media is the video component;
// Photo is the corresponding static image.
type InputMediaLivePhoto struct {
	Media                 string          `json:"media"`
	Photo                 string          `json:"photo"`
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	HasSpoiler            bool            `json:"has_spoiler,omitempty"`
}

func (InputMediaLivePhoto) mediaGroupItem()        {}
func (InputMediaLivePhoto) mediaGroupType() string { return "live_photo" }
func (m InputMediaLivePhoto) mediaGroupSource() string {
	return m.Media
}
func (InputMediaLivePhoto) inputPollMedia()       {}
func (InputMediaLivePhoto) inputPollOptionMedia() {}
func (m InputMediaLivePhoto) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != "" && strings.TrimSpace(m.Photo) != ""
}
func (m InputMediaLivePhoto) MarshalJSON() ([]byte, error) {
	type alias InputMediaLivePhoto
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "live_photo", alias: alias(m)})
}

// InputMediaLocation represents a static location in poll media.
type InputMediaLocation struct {
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	HorizontalAccuracy float64 `json:"horizontal_accuracy,omitempty"`
}

func (InputMediaLocation) inputPollMedia()       {}
func (InputMediaLocation) inputPollOptionMedia() {}
func (m InputMediaLocation) validPollMedia() bool {
	return m.Latitude >= -90 && m.Latitude <= 90 && m.Longitude >= -180 && m.Longitude <= 180
}
func (m InputMediaLocation) MarshalJSON() ([]byte, error) {
	type alias InputMediaLocation
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "location", alias: alias(m)})
}

// InputMediaVenue represents a venue in poll media.
type InputMediaVenue struct {
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Title           string  `json:"title"`
	Address         string  `json:"address"`
	FoursquareID    string  `json:"foursquare_id,omitempty"`
	FoursquareType  string  `json:"foursquare_type,omitempty"`
	GooglePlaceID   string  `json:"google_place_id,omitempty"`
	GooglePlaceType string  `json:"google_place_type,omitempty"`
}

func (InputMediaVenue) inputPollMedia()       {}
func (InputMediaVenue) inputPollOptionMedia() {}
func (m InputMediaVenue) validPollMedia() bool {
	return m.Latitude >= -90 && m.Latitude <= 90 && m.Longitude >= -180 && m.Longitude <= 180 &&
		strings.TrimSpace(m.Title) != "" && strings.TrimSpace(m.Address) != ""
}
func (m InputMediaVenue) MarshalJSON() ([]byte, error) {
	type alias InputMediaVenue
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "venue", alias: alias(m)})
}

// InputMediaLink represents an HTTP link on a poll option.
type InputMediaLink struct {
	URL string `json:"url"`
}

func (InputMediaLink) inputPollOptionMedia() {}
func (m InputMediaLink) validPollMedia() bool {
	value := strings.TrimSpace(m.URL)
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}
func (m InputMediaLink) MarshalJSON() ([]byte, error) {
	type alias InputMediaLink
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "link", alias: alias(m)})
}

// InputMediaSticker represents a sticker on a poll option.
type InputMediaSticker struct {
	Media string `json:"media"`
	Emoji string `json:"emoji,omitempty"`
}

func (InputMediaSticker) inputPollOptionMedia() {}
func (m InputMediaSticker) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != ""
}
func (m InputMediaSticker) MarshalJSON() ([]byte, error) {
	type alias InputMediaSticker
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "sticker", alias: alias(m)})
}

func (InputMediaPhoto) inputPollMedia()       {}
func (InputMediaPhoto) inputPollOptionMedia() {}
func (m InputMediaPhoto) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != ""
}

func (InputMediaVideo) inputPollMedia()       {}
func (InputMediaVideo) inputPollOptionMedia() {}
func (m InputMediaVideo) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != ""
}

func (InputMediaAudio) inputPollMedia() {}
func (m InputMediaAudio) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != ""
}

func (InputMediaDocument) inputPollMedia() {}
func (m InputMediaDocument) validPollMedia() bool {
	return strings.TrimSpace(m.Media) != ""
}
