package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestAttachmentScannerIgnoresEscapedQuotePrefix(t *testing.T) {
	t.Parallel()

	value := struct {
		Media   string `json:"media"`
		Caption string `json:"caption"`
	}{
		Media:   Attachment("required"),
		Caption: `literal "attach://not-an-upload" text`,
	}
	if err := validateAttachmentUploads(value, []Upload{
		NewUpload("required", "file.bin", strings.NewReader("data")),
	}, "test"); err != nil {
		t.Fatal(err)
	}
}

func TestAttachmentScannerUnquotesFieldName(t *testing.T) {
	t.Parallel()

	field := `quoted"field`
	if err := validateAttachmentUploads(Attachment(field), []Upload{
		NewUpload(field, "file.bin", strings.NewReader("data")),
	}, "test"); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkValidateAttachmentUploads(b *testing.B) {
	value := []InputPaidMedia{
		InputPaidMediaLivePhoto{Media: Attachment("video"), Photo: Attachment("photo")},
		InputPaidMediaVideo{Media: "existing-video"},
	}
	uploads := []Upload{
		NewUpload("video", "video.mp4", strings.NewReader("video")),
		NewUpload("photo", "photo.jpg", strings.NewReader("photo")),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := validateAttachmentUploads(value, uploads, "benchmark"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateAttachmentUploadsLegacy(b *testing.B) {
	value := []InputPaidMedia{
		InputPaidMediaLivePhoto{Media: Attachment("video"), Photo: Attachment("photo")},
		InputPaidMediaVideo{Media: "existing-video"},
	}
	uploads := []Upload{
		NewUpload("video", "video.mp4", strings.NewReader("video")),
		NewUpload("photo", "photo.jpg", strings.NewReader("photo")),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := validateAttachmentUploadsLegacy(value, uploads, "benchmark"); err != nil {
			b.Fatal(err)
		}
	}
}

// validateAttachmentUploadsLegacy preserves the pre-optimization generic JSON
// tree walk so the benchmark records a directly comparable baseline.
func validateAttachmentUploadsLegacy(value any, uploads []Upload, method string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var decoded any
	if err = json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	references := make(map[string]struct{})
	collectAttachmentReferencesLegacy(decoded, references)
	provided := make(map[string]struct{}, len(uploads))
	for _, upload := range uploads {
		field := strings.TrimSpace(upload.Field)
		if field == "" || upload.Reader == nil {
			return fmt.Errorf("invalid %s upload", method)
		}
		provided[field] = struct{}{}
	}
	for field := range references {
		if _, exists := provided[field]; !exists {
			return fmt.Errorf("missing %s upload", method)
		}
	}
	for field := range provided {
		if _, exists := references[field]; !exists {
			return fmt.Errorf("extra %s upload", method)
		}
	}
	return nil
}

func collectAttachmentReferencesLegacy(value any, fields map[string]struct{}) {
	switch current := value.(type) {
	case string:
		if strings.HasPrefix(current, "attach://") {
			field := strings.TrimSpace(strings.TrimPrefix(current, "attach://"))
			if field != "" {
				fields[field] = struct{}{}
			}
		}
	case []any:
		for _, item := range current {
			collectAttachmentReferencesLegacy(item, fields)
		}
	case map[string]any:
		for _, item := range current {
			collectAttachmentReferencesLegacy(item, fields)
		}
	}
}

func BenchmarkValidMethod(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if !validMethod("sendRichMessageDraft") {
			b.Fatal("valid method rejected")
		}
	}
}
