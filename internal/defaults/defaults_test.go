package defaults

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticUTIsKeysStartWithDot(t *testing.T) {
	for ext := range staticUTIs {
		if !strings.HasPrefix(ext, ".") {
			t.Errorf("staticUTIs key %q does not start with '.'", ext)
		}
	}
}

func TestStaticUTIsValuesNonEmpty(t *testing.T) {
	for ext, uti := range staticUTIs {
		if uti == "" {
			t.Errorf("staticUTIs[%q] is empty", ext)
		}
	}
}

func TestStaticUTIsKnownMappings(t *testing.T) {
	cases := map[string]string{
		".json": "public.json",
		".py":   "public.python-script",
		".html": "public.html",
		".pdf":  "com.adobe.pdf",
		".png":  "public.png",
	}
	for ext, want := range cases {
		got, ok := staticUTIs[ext]
		if !ok {
			t.Errorf("staticUTIs missing %q", ext)
		} else if got != want {
			t.Errorf("staticUTIs[%q] = %q, want %q", ext, got, want)
		}
	}
}

func TestUpsertByUTIInsert(t *testing.T) {
	db := &lsDatabase{}
	upsertByUTI(db, "public.json", "com.example.app", 100)
	if len(db.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(db.Handlers))
	}
	h := db.Handlers[0]
	if h.ContentType != "public.json" || h.RoleAll != "com.example.app" {
		t.Errorf("handler = %+v", h)
	}
}

func TestUpsertByUTIUpdate(t *testing.T) {
	db := &lsDatabase{
		Handlers: []lsHandler{{ContentType: "public.json", RoleAll: "com.old.app"}},
	}
	upsertByUTI(db, "public.json", "com.new.app", 200)
	if len(db.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(db.Handlers))
	}
	if db.Handlers[0].RoleAll != "com.new.app" {
		t.Errorf("RoleAll = %q, want com.new.app", db.Handlers[0].RoleAll)
	}
}

func TestUpsertByTagInsert(t *testing.T) {
	db := &lsDatabase{}
	upsertByTag(db, "vue", "com.example.app", 100)
	if len(db.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(db.Handlers))
	}
	h := db.Handlers[0]
	if h.ContentTag != "vue" || h.ContentTagClass != "public.filename-extension" || h.RoleAll != "com.example.app" {
		t.Errorf("handler = %+v", h)
	}
}

func TestUpsertByTagUpdate(t *testing.T) {
	db := &lsDatabase{
		Handlers: []lsHandler{{
			ContentTag:      "vue",
			ContentTagClass: "public.filename-extension",
			RoleAll:         "com.old.app",
		}},
	}
	upsertByTag(db, "vue", "com.new.app", 200)
	if len(db.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(db.Handlers))
	}
	if db.Handlers[0].RoleAll != "com.new.app" {
		t.Errorf("RoleAll = %q, want com.new.app", db.Handlers[0].RoleAll)
	}
}

func TestUpsertURLSchemeInsert(t *testing.T) {
	db := &lsDatabase{}
	upsertURLScheme(db, "https", "com.example.browser", 100)
	if len(db.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(db.Handlers))
	}
	if db.Handlers[0].URLScheme != "https" || db.Handlers[0].RoleAll != "com.example.browser" {
		t.Errorf("handler = %+v", db.Handlers[0])
	}
}

func TestUpsertURLSchemeUpdate(t *testing.T) {
	db := &lsDatabase{
		Handlers: []lsHandler{{URLScheme: "https", RoleAll: "com.old.browser"}},
	}
	upsertURLScheme(db, "https", "com.new.browser", 200)
	if len(db.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(db.Handlers))
	}
	if db.Handlers[0].RoleAll != "com.new.browser" {
		t.Errorf("RoleAll = %q, want com.new.browser", db.Handlers[0].RoleAll)
	}
}

func TestPruneBackups(t *testing.T) {
	dir := t.TempDir()
	// Create 5 backup files
	for i := 0; i < 5; i++ {
		name := filepath.Join(dir, "launchservices-2024010"+string(rune('0'+i))+"-120000.plist")
		if err := os.WriteFile(name, []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	pruneBackups(dir, 3)
	entries, _ := os.ReadDir(dir)
	if len(entries) != 3 {
		t.Errorf("expected 3 backups after prune, got %d", len(entries))
	}
}

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	data := []byte("hello world")

	if err := atomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}

	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0644 {
		t.Errorf("permissions = %o, want 0644", info.Mode().Perm())
	}
}

func TestValidBundleIDPattern(t *testing.T) {
	valid := []string{"com.apple.Safari", "dev.zed.Zed", "com.microsoft.VSCode"}
	invalid := []string{"com.apple.Safari;rm -rf", "'; DROP TABLE", "com.app/bad"}

	for _, v := range valid {
		if !ValidBundleIDPattern.MatchString(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	for _, v := range invalid {
		if ValidBundleIDPattern.MatchString(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}
