		t.Fatal(err)
	}
	if schema.Version != "10.5" || schema.Released != "November 22, 2026" || schema.Source != "https://core.telegram.org/bots/api" {
		t.Fatalf("metadata = %#v", schema)
	}
	want := Stats{Methods: 194, Parameters: 997, Objects: 397, ObjectFields: 1992, Unions: 29, Variants: 204}
	if got := schema.Stats(); got != want {
		t.Fatalf("stats = %#v, want %#v", got, want)
	}
}
