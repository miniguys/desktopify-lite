package app

import (
	"bufio"
	"strings"
	"testing"
)

func TestResolveRunInputTracksExplicitIconURL(t *testing.T) {
	in, err := ResolveRunInput(RuntimeOptions{
		NonInteractive: true,
		RunInput: Input{
			URL:     "https://example.com",
			Name:    "Example",
			IconURL: "https://example.com/icon?id=123",
		},
		RunInputExplicit: RunInputExplicit{IconURL: true},
	}, DefaultConfig(), bufio.NewReader(strings.NewReader("")))
	if err != nil {
		t.Fatal(err)
	}
	if !in.IconURLExplicit {
		t.Fatal("expected explicit icon URL to be preserved")
	}
}

func TestResolveRunInputSkipIconLeavesIconURLBlank(t *testing.T) {
	in, err := ResolveRunInput(RuntimeOptions{
		NonInteractive: true,
		SkipIcon:       true,
		RunInput: Input{
			URL:  "https://example.com",
			Name: "Example",
		},
	}, DefaultConfig(), bufio.NewReader(strings.NewReader("")))
	if err != nil {
		t.Fatal(err)
	}
	if in.IconURL != "" {
		t.Fatalf("IconURL=%q, want empty", in.IconURL)
	}
}
