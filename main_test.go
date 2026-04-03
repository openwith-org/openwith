package main

import (
	"strings"
	"testing"
)

func TestParseImportLinesValid(t *testing.T) {
	input := ".json = com.microsoft.VSCode  # VS Code\n.py = dev.zed.Zed\n"
	changes, err := parseImportLines(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[".json"] != "com.microsoft.VSCode" {
		t.Errorf(".json = %q", changes[".json"])
	}
	if changes[".py"] != "dev.zed.Zed" {
		t.Errorf(".py = %q", changes[".py"])
	}
}

func TestParseImportLinesSkipsComments(t *testing.T) {
	input := "# header comment\n\n.go = dev.zed.Zed\n# another comment\n"
	changes, err := parseImportLines(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestParseImportLinesEmpty(t *testing.T) {
	input := "# only comments\n\n"
	_, err := parseImportLines(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for empty result")
	}
}

func TestParseImportLinesStripsInlineComments(t *testing.T) {
	input := ".rs = dev.zed.Zed  # Zed editor\n"
	changes, err := parseImportLines(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changes[".rs"] != "dev.zed.Zed" {
		t.Errorf(".rs = %q, want dev.zed.Zed", changes[".rs"])
	}
}
