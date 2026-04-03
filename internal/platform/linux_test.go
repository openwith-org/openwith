//go:build linux

package platform

import (
	"testing"
)

func TestMimeTypeForExtKnown(t *testing.T) {
	cases := map[string]string{
		".json": "application/json",
		".py":   "text/x-python",
		".html": "text/html",
		".pdf":  "application/pdf",
		".png":  "image/png",
	}
	for ext, want := range cases {
		got := mimeTypeForExt(ext)
		if got != want {
			t.Errorf("mimeTypeForExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestMimeTypeForExtFallback(t *testing.T) {
	got := mimeTypeForExt(".unknownext")
	if got != "text/plain" {
		t.Errorf("mimeTypeForExt(.unknownext) = %q, want text/plain", got)
	}
}
