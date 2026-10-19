package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEditorialPack(t *testing.T) {
	read := func(name string, out any) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("data", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatal(err)
		}
	}
	var pack contentPack
	var credits []photoCredit
	read("content.json", &pack)
	read("photo-credits.json", &credits)
	byKey, err := validate(pack, credits, "https://tuyou.evenstar.love/editorial/")
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Posts) != 33 || len(byKey) != 25 {
		t.Fatalf("unexpected content size: %d posts, %d photos", len(pack.Posts), len(byKey))
	}
	for _, p := range pack.Posts {
		credit := byKey[p.Photo]
		if credit.Author == "" || credit.Source == "" || credit.LicenseURL == "" {
			t.Fatalf("missing photo attribution for %s", p.Title)
		}
	}
}
