package extensions

import (
	"testing"
)

func TestAllExtsCount(t *testing.T) {
	exts := AllExts()
	if len(exts) != len(All) {
		t.Errorf("AllExts() returned %d, want %d", len(exts), len(All))
	}
}

func TestAllExtsStartWithDot(t *testing.T) {
	for _, ext := range AllExts() {
		if ext[0] != '.' {
			t.Errorf("extension %q does not start with '.'", ext)
		}
	}
}

func TestAllExtsNoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, e := range All {
		if seen[e.Ext] {
			t.Errorf("duplicate extension: %s", e.Ext)
		}
		seen[e.Ext] = true
	}
}

func TestGroupedCoversAllExtensions(t *testing.T) {
	groups := Grouped()
	total := 0
	for _, g := range groups {
		total += len(g.Extensions)
	}
	if total != len(All) {
		t.Errorf("Grouped() has %d total extensions, want %d", total, len(All))
	}
}

func TestGroupedFollowsCategoryOrder(t *testing.T) {
	groups := Grouped()
	orderIdx := make(map[Category]int)
	for i, c := range CategoryOrder {
		orderIdx[c] = i
	}
	for i := 1; i < len(groups); i++ {
		prev := orderIdx[groups[i-1].Category]
		curr := orderIdx[groups[i].Category]
		if curr <= prev {
			t.Errorf("group %q (index %d) appears before %q (index %d) but should be after",
				groups[i].Category, curr, groups[i-1].Category, prev)
		}
	}
}

func TestMergeAddsNew(t *testing.T) {
	// Save and restore All
	orig := make([]Extension, len(All))
	copy(orig, All)
	defer func() { All = orig }()

	before := len(All)
	Merge([]Extension{{".testonly", Other}})
	if len(All) != before+1 {
		t.Errorf("Merge should add 1 extension, got %d (was %d)", len(All), before)
	}
}

func TestMergeSkipsDuplicates(t *testing.T) {
	orig := make([]Extension, len(All))
	copy(orig, All)
	defer func() { All = orig }()

	before := len(All)
	Merge([]Extension{{".json", Config}})
	if len(All) != before {
		t.Errorf("Merge should skip duplicate, got %d (was %d)", len(All), before)
	}
}
