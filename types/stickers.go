package types

type MaskPosition struct {
	Point  string  `json:"point"`
	XShift float64 `json:"x_shift"`
	YShift float64 `json:"y_shift"`
	Scale  float64 `json:"scale"`
}

type StickerSet struct {
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	StickerType string     `json:"sticker_type"`
	Stickers    []Sticker  `json:"stickers"`
	Thumbnail   *PhotoSize `json:"thumbnail,omitempty"`
}
