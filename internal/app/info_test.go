package app

import (
	"strings"
	"testing"
)

func TestVersionLineIncludesVersion(t *testing.T) {
	got := versionLine()
	if !strings.HasPrefix(got, "desktopify-lite ") {
		t.Fatalf("versionLine=%q", got)
	}
	if !strings.Contains(got, version) {
		t.Fatalf("versionLine should include version %q, got %q", version, got)
	}
}
