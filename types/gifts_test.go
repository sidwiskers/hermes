package types

import (
	"encoding/json"
	"testing"
)

func TestOwnedGiftDiscriminatorDecode(t *testing.T) {
	t.Parallel()

	var gifts OwnedGifts
	err := json.Unmarshal([]byte(`{"total_count":2,"gifts":[{"type":"regular","gift":{"id":"regular-1","sticker":{"file_id":"s","file_unique_id":"u","type":"regular","width":1,"height":1,"is_animated":false,"is_video":false},"star_count":25},"send_date":1},{"type":"unique","gift":{"gift_id":"regular-1","base_name":"Hermes","name":"Hermes-1","number":1,"model":{},"symbol":{},"backdrop":{}},"send_date":2}]}`), &gifts)
	if err != nil {
		t.Fatal(err)
	}
	if gifts.Gifts[0].Gift == nil || gifts.Gifts[0].Gift.ID != "regular-1" {
		t.Fatalf("regular gift = %#v", gifts.Gifts[0])
	}
	if gifts.Gifts[1].UniqueGift == nil || gifts.Gifts[1].UniqueGift.Name != "Hermes-1" {
		t.Fatalf("unique gift = %#v", gifts.Gifts[1])
	}
}
