package app

import "testing"

func TestEmbeddedLogoExists(t *testing.T) {
	if embeddedLogo == "" {
		t.Fatal("embedded logo should not be empty")
	}
}
