package types

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ChatShareTheme is Telegram's chat-share theme union. Unknown variants are
// retained in Raw so newer Bot API responses can be inspected and re-emitted.
type ChatShareTheme struct {
	Emoji      *ChatShareThemeEmoji
	UniqueGift *ChatShareThemeUniqueGift
	Raw        json.RawMessage
}

func NewChatShareThemeEmoji(emojis []EmojiType, backgroundBackdropID int) ChatShareTheme {
	return ChatShareTheme{Emoji: &ChatShareThemeEmoji{
		Emojis: emojis, BackgroundBackdropID: backgroundBackdropID,
	}}
}

func NewChatShareThemeUniqueGift(gift UniqueGift) ChatShareTheme {
	return ChatShareTheme{UniqueGift: &ChatShareThemeUniqueGift{Gift: gift}}
}

func (value ChatShareTheme) Kind() string {
	switch {
	case value.Emoji != nil:
		return "emoji"
	case value.UniqueGift != nil:
		return "unique_gift"
	default:
		return ""
	}
}

func (value ChatShareTheme) MarshalJSON() ([]byte, error) {
	switch {
	case value.Emoji != nil && value.UniqueGift == nil:
		return json.Marshal(value.Emoji)
	case value.UniqueGift != nil && value.Emoji == nil:
		return json.Marshal(value.UniqueGift)
	case value.Emoji != nil || value.UniqueGift != nil:
		return nil, fmt.Errorf("hermes: ChatShareTheme must contain exactly one variant")
	case len(bytes.TrimSpace(value.Raw)) != 0:
		return bytes.Clone(value.Raw), nil
	default:
		return []byte("null"), nil
	}
}

func (value *ChatShareTheme) UnmarshalJSON(data []byte) error {
	if value == nil {
		return fmt.Errorf("hermes: cannot decode ChatShareTheme into a nil receiver")
	}
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	value.Emoji = nil
	value.UniqueGift = nil
	value.Raw = nil
	switch probe.Type {
	case "emoji":
		value.Emoji = new(ChatShareThemeEmoji)
		return json.Unmarshal(data, value.Emoji)
	case "unique_gift":
		value.UniqueGift = new(ChatShareThemeUniqueGift)
		return json.Unmarshal(data, value.UniqueGift)
	default:
		value.Raw = bytes.Clone(data)
		return nil
	}
}

func (value ChatShareThemeEmoji) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("emoji", struct {
		Emojis               []EmojiType `json:"emojis"`
		BackgroundBackdropID int         `json:"background_backdrop_id,omitempty"`
	}{value.Emojis, value.BackgroundBackdropID})
}

func (value ChatShareThemeUniqueGift) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("unique_gift", struct {
		Gift UniqueGift `json:"gift"`
	}{value.Gift})
}

// EmojiType is Telegram's emoji representation union.
type EmojiType struct {
	Emoji       *EmojiTypeEmoji
	CustomEmoji *EmojiTypeCustomEmoji
	Raw         json.RawMessage
}

func NewEmoji(value string) EmojiType {
	return EmojiType{Emoji: &EmojiTypeEmoji{Emoji: value}}
}

func NewCustomEmoji(customEmojiID string) EmojiType {
	return EmojiType{CustomEmoji: &EmojiTypeCustomEmoji{CustomEmojiID: customEmojiID}}
}

func (value EmojiType) Kind() string {
	switch {
	case value.Emoji != nil:
		return "emoji"
	case value.CustomEmoji != nil:
		return "custom_emoji"
	default:
		return ""
	}
}

func (value EmojiType) MarshalJSON() ([]byte, error) {
	switch {
	case value.Emoji != nil && value.CustomEmoji == nil:
		return json.Marshal(value.Emoji)
	case value.CustomEmoji != nil && value.Emoji == nil:
		return json.Marshal(value.CustomEmoji)
	case value.Emoji != nil || value.CustomEmoji != nil:
		return nil, fmt.Errorf("hermes: EmojiType must contain exactly one variant")
	case len(bytes.TrimSpace(value.Raw)) != 0:
		return bytes.Clone(value.Raw), nil
	default:
		return []byte("null"), nil
	}
}

func (value *EmojiType) UnmarshalJSON(data []byte) error {
	if value == nil {
		return fmt.Errorf("hermes: cannot decode EmojiType into a nil receiver")
	}
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	value.Emoji = nil
	value.CustomEmoji = nil
	value.Raw = nil
	switch probe.Type {
	case "emoji":
		value.Emoji = new(EmojiTypeEmoji)
		return json.Unmarshal(data, value.Emoji)
	case "custom_emoji":
		value.CustomEmoji = new(EmojiTypeCustomEmoji)
		return json.Unmarshal(data, value.CustomEmoji)
	default:
		value.Raw = bytes.Clone(data)
		return nil
	}
}

func (value EmojiTypeEmoji) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("emoji", struct {
		Emoji string `json:"emoji"`
	}{value.Emoji})
}

func (value EmojiTypeCustomEmoji) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("custom_emoji", struct {
		CustomEmojiID string `json:"custom_emoji_id"`
	}{value.CustomEmojiID})
}

// WidgetQueryType is Telegram's widget-query event union.
type WidgetQueryType struct {
	Data            *WidgetQueryTypeData
	SignData        *WidgetQueryTypeSignData
	SentTransaction *WidgetQueryTypeSentTransaction
	Raw             json.RawMessage
}

func NewWidgetQueryData(walletCard *TonWalletCard) WidgetQueryType {
	return WidgetQueryType{Data: &WidgetQueryTypeData{WalletCard: walletCard}}
}

func NewWidgetQuerySignData(wallet WidgetQueryWallet, signData string) WidgetQueryType {
	return WidgetQueryType{SignData: &WidgetQueryTypeSignData{Wallet: wallet, SignData: signData}}
}

func NewWidgetQuerySentTransaction(wallet WidgetQueryWallet, msgHash string) WidgetQueryType {
	return WidgetQueryType{SentTransaction: &WidgetQueryTypeSentTransaction{Wallet: wallet, MsgHash: msgHash}}
}

func (value WidgetQueryType) Kind() string {
	switch {
	case value.Data != nil:
		return "data"
	case value.SignData != nil:
		return "sign_data"
	case value.SentTransaction != nil:
		return "sent_transaction"
	default:
		return ""
	}
}

func (value WidgetQueryType) MarshalJSON() ([]byte, error) {
	switch {
	case value.Data != nil && value.SignData == nil && value.SentTransaction == nil:
		return json.Marshal(value.Data)
	case value.SignData != nil && value.Data == nil && value.SentTransaction == nil:
		return json.Marshal(value.SignData)
	case value.SentTransaction != nil && value.Data == nil && value.SignData == nil:
		return json.Marshal(value.SentTransaction)
	case value.Data != nil || value.SignData != nil || value.SentTransaction != nil:
		return nil, fmt.Errorf("hermes: WidgetQueryType must contain exactly one variant")
	case len(bytes.TrimSpace(value.Raw)) != 0:
		return bytes.Clone(value.Raw), nil
	default:
		return []byte("null"), nil
	}
}

func (value *WidgetQueryType) UnmarshalJSON(data []byte) error {
	if value == nil {
		return fmt.Errorf("hermes: cannot decode WidgetQueryType into a nil receiver")
	}
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	value.Data = nil
	value.SignData = nil
	value.SentTransaction = nil
	value.Raw = nil
	switch probe.Type {
	case "data":
		value.Data = new(WidgetQueryTypeData)
		return json.Unmarshal(data, value.Data)
	case "sign_data":
		value.SignData = new(WidgetQueryTypeSignData)
		return json.Unmarshal(data, value.SignData)
	case "sent_transaction":
		value.SentTransaction = new(WidgetQueryTypeSentTransaction)
		return json.Unmarshal(data, value.SentTransaction)
	default:
		value.Raw = bytes.Clone(data)
		return nil
	}
}

func (value WidgetQueryTypeData) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("data", struct {
		WalletCard *TonWalletCard `json:"wallet_card,omitempty"`
	}{value.WalletCard})
}

func (value WidgetQueryTypeSignData) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("sign_data", struct {
		Wallet   WidgetQueryWallet `json:"wallet"`
		SignData string            `json:"sign_data"`
	}{value.Wallet, value.SignData})
}

func (value WidgetQueryTypeSentTransaction) MarshalJSON() ([]byte, error) {
	return marshalDiscriminated("sent_transaction", struct {
		Wallet  WidgetQueryWallet `json:"wallet"`
		MsgHash string            `json:"msg_hash"`
	}{value.Wallet, value.MsgHash})
}

// The two Bot API 10.4 widget message-content objects are valid members of
// the inline-query InputMessageContent union.
func (InputWidgetMessageContent) inputMessageContent()       {}
func (InputSharedWidgetMessageContent) inputMessageContent() {}
