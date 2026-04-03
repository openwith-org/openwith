package editors

import (
	"testing"
)

func TestSortByPriorityEmpty(t *testing.T) {
	eds := []Editor{{Name: "A", BundleID: "a"}, {Name: "B", BundleID: "b"}}
	result := SortByPriority(eds, nil)
	if len(result) != 2 || result[0].Name != "A" {
		t.Errorf("empty priority should return input unchanged")
	}
}

func TestSortByPriorityReorders(t *testing.T) {
	eds := []Editor{
		{Name: "A", BundleID: "a"},
		{Name: "B", BundleID: "b"},
		{Name: "C", BundleID: "c"},
	}
	result := SortByPriority(eds, []string{"c", "a"})
	if result[0].BundleID != "c" || result[1].BundleID != "a" || result[2].BundleID != "b" {
		t.Errorf("got order %s,%s,%s; want c,a,b", result[0].BundleID, result[1].BundleID, result[2].BundleID)
	}
}

func TestSortByPriorityUnmatchedAppearAfter(t *testing.T) {
	eds := []Editor{
		{Name: "X", BundleID: "x"},
		{Name: "Y", BundleID: "y"},
	}
	result := SortByPriority(eds, []string{"y"})
	if result[0].BundleID != "y" || result[1].BundleID != "x" {
		t.Errorf("got %s,%s; want y,x", result[0].BundleID, result[1].BundleID)
	}
}

func TestSortByPriorityEmptyEditors(t *testing.T) {
	result := SortByPriority(nil, []string{"a"})
	if len(result) != 0 {
		t.Errorf("empty editors should return empty, got %d", len(result))
	}
}

func TestRegistryIntegrity(t *testing.T) {
	allEditors := make([]Editor, 0, len(Registry)+len(MediaApps)+len(Browsers))
	allEditors = append(allEditors, Registry...)
	allEditors = append(allEditors, MediaApps...)
	allEditors = append(allEditors, Browsers...)

	keys := make(map[string]bool)
	bundles := make(map[string]bool)
	for _, ed := range allEditors {
		if ed.Name == "" {
			t.Error("editor has empty Name")
		}
		if ed.Key == "" {
			t.Errorf("editor %q has empty Key", ed.Name)
		}
		if ed.BundleID == "" {
			t.Errorf("editor %q has empty BundleID", ed.Name)
		}
		if keys[ed.Key] {
			t.Errorf("duplicate key: %s", ed.Key)
		}
		keys[ed.Key] = true
		if bundles[ed.BundleID] {
			t.Errorf("duplicate bundle ID: %s", ed.BundleID)
		}
		bundles[ed.BundleID] = true
	}
}
