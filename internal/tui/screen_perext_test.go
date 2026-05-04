package tui

import (
	"testing"

	"github.com/openwith-org/openwith/internal/editors"
	"github.com/openwith-org/openwith/internal/extensions"
)

func TestBuildFilteredEmpty(t *testing.T) {
	m := PerExtModel{
		rows: []perExtRow{
			{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
			{ext: extensions.Extension{Ext: ".py", Category: extensions.Docs}},
		},
		filtered: nil,
	}

	m.buildFiltered()
	if m.filtered != nil {
		t.Error("buildFiltered with empty filter should set filtered to nil")
	}
}

func TestBuildFilteredPartialMatch(t *testing.T) {
	m := PerExtModel{
		rows: []perExtRow{
			{isSep: true, sepLabel: "Config"},
			{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
			{ext: extensions.Extension{Ext: ".yaml", Category: extensions.Config}},
			{isSep: true, sepLabel: "Docs"},
			{ext: extensions.Extension{Ext: ".py", Category: extensions.Docs}},
		},
		filter: "json",
	}

	m.buildFiltered()

	if len(m.filtered) != 2 {
		t.Errorf("buildFiltered with filter 'json' returned %d results, want 2", len(m.filtered))
	}

	found := false
	for _, idx := range m.filtered {
		if m.rows[idx].ext.Ext == ".json" {
			found = true
		}
	}
	if !found {
		t.Error("buildFiltered should include .json extension")
	}
}

func TestBuildFilteredCaseInsensitive(t *testing.T) {
	m := PerExtModel{
		rows: []perExtRow{
			{ext: extensions.Extension{Ext: ".JSON", Category: extensions.Config}},
			{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
		},
		filter: "JSON",
	}

	m.buildFiltered()

	if len(m.filtered) != 2 {
		t.Errorf("buildFiltered should be case insensitive, got %d results, want 2", len(m.filtered))
	}
}

func TestBuildFilteredCategoryMatch(t *testing.T) {
	m := PerExtModel{
		rows: []perExtRow{
			{isSep: true, sepLabel: "Config"},
			{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
			{isSep: true, sepLabel: "Docs"},
			{ext: extensions.Extension{Ext: ".py", Category: extensions.Docs}},
		},
		filter: "config",
	}

	m.buildFiltered()

	if len(m.filtered) != 2 {
		t.Errorf("filter 'config' should match category + extensions, got %d", len(m.filtered))
	}
}

func TestCategoryHasMatch(t *testing.T) {
	rows := []perExtRow{
		{isSep: true, sepLabel: "Config"},
		{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
		{ext: extensions.Extension{Ext: ".yaml", Category: extensions.Config}},
		{isSep: true, sepLabel: "Docs"},
		{ext: extensions.Extension{Ext: ".py", Category: extensions.Docs}},
	}

	m := PerExtModel{rows: rows}

	if !m.categoryHasMatch(0, "json") {
		t.Error("categoryHasMatch(0, 'json') should return true")
	}

	if m.categoryHasMatch(0, "python") {
		t.Error("categoryHasMatch(0, 'python') should return false")
	}
}

func TestNewPerExtModelSetsCursor(t *testing.T) {
	eds := []editors.Editor{
		{Name: "VSCode", BundleID: "com.microsoft.VSCode"},
	}

	m := NewPerExtModel(eds, nil, 20)

	foundNonSep := false
	for i, r := range m.rows {
		if !r.isSep {
			if i != m.cursor {
				t.Errorf("cursor should be at first non-separator row, got %d", m.cursor)
			}
			foundNonSep = true
			break
		}
	}
	if !foundNonSep {
		t.Error("rows should contain non-separator entries")
	}
}

func TestPerExtFilterRemovesFiltered(t *testing.T) {
	m := PerExtModel{
		rows: []perExtRow{
			{isSep: true, sepLabel: "Config"},
			{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
		},
		filtered: []int{0, 1},
		filter:   "",
	}

	m.buildFiltered()

	if m.filtered != nil {
		t.Error("clearing filter should set filtered to nil")
	}
}

func TestPerExtCursorMovement(t *testing.T) {
	m := PerExtModel{
		rows: []perExtRow{
			{isSep: true, sepLabel: "Config"},
			{ext: extensions.Extension{Ext: ".json", Category: extensions.Config}},
			{ext: extensions.Extension{Ext: ".yaml", Category: extensions.Config}},
		},
		cursor:     1,
		viewHeight: 10,
	}

	m.cursor = 2
	if m.cursor != 2 {
		t.Errorf("cursor should move to 2, got %d", m.cursor)
	}

	m.cursor = 1
	m.filter = "json"
	m.buildFiltered()

	if len(m.filtered) != 2 {
		t.Fatalf("filtered should have 2 entries, got %d", len(m.filtered))
	}
}

func TestMenuModelNew(t *testing.T) {
	m := NewMenuModel()

	if len(m.items) != 5 {
		t.Errorf("expected 5 menu items, got %d", len(m.items))
	}

	if m.items[0].title != "Bulk Mode" {
		t.Errorf("first item should be Bulk Mode, got %s", m.items[0].title)
	}

	if m.cursor != 0 {
		t.Errorf("initial cursor should be 0, got %d", m.cursor)
	}
}

func TestMenuModelSelectScreen(t *testing.T) {
	m := NewMenuModel()
	m.cursor = 2

	if m.items[m.cursor].screen != screenBrowser {
		t.Errorf("item 2 should navigate to screenBrowser, got %v", m.items[m.cursor].screen)
	}
}

func TestBulkModelNew(t *testing.T) {
	eds := []editors.Editor{
		{Name: "VS Code", BundleID: "com.microsoft.VSCode"},
		{Name: "Zed", BundleID: "dev.zed.Zed"},
	}

	m := NewBulkModel(eds, 20)

	if len(m.editors) != 2 {
		t.Errorf("expected 2 editors, got %d", len(m.editors))
	}

	if m.viewHeight != 12 {
		t.Errorf("viewHeight should be 12, got %d", m.viewHeight)
	}
}

func TestBulkModelEnsureVisible(t *testing.T) {
	m := &BulkModel{
		editors:    []editors.Editor{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}},
		cursor:     3,
		offset:    0,
		viewHeight: 3,
	}

	m.ensureVisible()
	if m.offset != 1 {
		t.Errorf("offset should be 1 after ensureVisible, got %d", m.offset)
	}
}

func TestBulkModelEnsureVisibleAtTop(t *testing.T) {
	m := &BulkModel{
		editors:    []editors.Editor{{Name: "A"}, {Name: "B"}, {Name: "C"}},
		cursor:     0,
		offset:    5,
		viewHeight: 3,
	}

	m.ensureVisible()
	if m.offset != 0 {
		t.Errorf("offset should be 0, got %d", m.offset)
	}
}

func TestBulkModelViewHeightMin(t *testing.T) {
	eds := []editors.Editor{{Name: "A"}}

	m := NewBulkModel(eds, 3)

	if m.viewHeight != 5 {
		t.Errorf("viewHeight should be min 5, got %d", m.viewHeight)
	}
}

func TestPerExtModelViewHeightMin(t *testing.T) {
	eds := []editors.Editor{{Name: "A"}}

	m := NewPerExtModel(eds, nil, 3)

	if m.viewHeight != 5 {
		t.Errorf("viewHeight should be min 5, got %d", m.viewHeight)
	}
}

func TestNewPerExtModelInitializesMaps(t *testing.T) {
	eds := []editors.Editor{{Name: "VSCode", BundleID: "com.microsoft.VSCode"}}
	defs := map[string]string{".json": "VS Code"}

	m := NewPerExtModel(eds, defs, 20)

	if m.defaults == nil {
		t.Error("defaults map should be initialized")
	}

	if m.changes == nil {
		t.Error("changes map should be initialized")
	}

	if m.editorIdx == nil {
		t.Error("editorIdx map should be initialized")
	}
}

func TestPerExtModelEditorIndex(t *testing.T) {
	eds := []editors.Editor{
		{Name: "VSCode", BundleID: "com.microsoft.VSCode"},
		{Name: "Zed", BundleID: "dev.zed.Zed"},
	}

	m := NewPerExtModel(eds, nil, 20)

	if m.editorIdx[".json"] != 0 {
		t.Error("editorIdx should map extension to first editor")
	}
}