package botapi

import (
	"reflect"
	"testing"
)

func TestCompareSurfaceIgnoresDocumentMetadata(t *testing.T) {
	before := testManifest()
	after := before
	after.Released = "tomorrow"
	after.SourceSHA256 = "different"

	diff := CompareSurface(before, after)
	if diff.Changed {
		t.Fatalf("metadata-only change reported as protocol change: %+v", diff)
	}
	if diff.Classification != "unchanged" {
		t.Fatalf("classification = %q, want unchanged", diff.Classification)
	}
}

func TestCompareSurfaceClassifiesProtocolChanges(t *testing.T) {
	before := testManifest()
	after := testManifest()
	after.Version = "2.0"
	after.Methods[0].Parameters = append(after.Methods[0].Parameters, Field{Name: "silent", Type: "Boolean"})
	after.Methods = append(after.Methods, Method{Name: "removeMessage", Anchor: "removemessage"})
	after.Objects = nil
	after.Unions[0].Variants = append(after.Unions[0].Variants, "Video")

	diff := CompareSurface(before, after)
	if !diff.Changed || !diff.VersionChanged {
		t.Fatalf("expected a versioned protocol change: %+v", diff)
	}
	if !reflect.DeepEqual(diff.Methods.Added, []string{"removeMessage"}) ||
		!reflect.DeepEqual(diff.Methods.Changed, []string{"sendMessage"}) {
		t.Fatalf("unexpected method classification: %+v", diff.Methods)
	}
	if !reflect.DeepEqual(diff.Objects.Removed, []string{"Message"}) {
		t.Fatalf("unexpected object classification: %+v", diff.Objects)
	}
	if !reflect.DeepEqual(diff.Unions.Changed, []string{"Media"}) {
		t.Fatalf("unexpected union classification: %+v", diff.Unions)
	}
	if diff.Classification != "review" || len(diff.ReviewReasons) == 0 {
		t.Fatalf("classification = %q, reasons = %v, want review", diff.Classification, diff.ReviewReasons)
	}
}

func TestCompareSurfaceClassifiesAdditiveObjectsAsMechanical(t *testing.T) {
	before := testManifest()
	after := testManifest()
	after.Version = "1.1"
	after.Objects[0].Fields = append(after.Objects[0].Fields, Field{Name: "text", Type: "String"})
	after.Objects = append(after.Objects, Object{Name: "Photo", Anchor: "photo"})

	diff := CompareSurface(before, after)
	if diff.Classification != "mechanical" || len(diff.ReviewReasons) != 0 {
		t.Fatalf("classification = %q, reasons = %v, want mechanical", diff.Classification, diff.ReviewReasons)
	}
}

func TestCompareSurfaceRequiresReviewForNewRequiredObjectField(t *testing.T) {
	before := testManifest()
	after := testManifest()
	after.Version = "1.1"
	after.Objects[0].Fields = append(after.Objects[0].Fields, Field{
		Name:     "business_connection_id",
		Type:     "String",
		Required: true,
	})

	diff := CompareSurface(before, after)
	if diff.Classification != "review" || len(diff.ReviewReasons) == 0 {
		t.Fatalf("classification = %q, reasons = %v, want review", diff.Classification, diff.ReviewReasons)
	}
}

func TestCompareSurfaceRequiresReviewForVersionOnlyRelease(t *testing.T) {
	before := testManifest()
	after := testManifest()
	after.Version = "1.1"

	diff := CompareSurface(before, after)
	if diff.Classification != "review" {
		t.Fatalf("classification = %q, want review", diff.Classification)
	}
}

func testManifest() Manifest {
	return Manifest{
		Version:      "1.0",
		Released:     "today",
		Source:       OfficialSource,
		SourceSHA256: "sha",
		Methods: []Method{{
			Name:   "sendMessage",
			Anchor: "sendmessage",
			Parameters: []Field{{
				Name:     "chat_id",
				Type:     "Integer",
				Required: true,
			}},
		}},
		Objects: []Object{{
			Name:   "Message",
			Anchor: "message",
			Fields: []Field{{
				Name:     "message_id",
				Type:     "Integer",
				Required: true,
			}},
		}},
		Unions: []Union{{
			Name:     "Media",
			Anchor:   "media",
			Variants: []string{"Photo"},
		}},
	}
}
