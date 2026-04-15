package goversion_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	goversion "github.com/cumulus13/go-version"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %q: %v", path, err)
	}
	return path
}

// ---------------------------------------------------------------------------
// ParseContent
// ---------------------------------------------------------------------------

func TestParseContent_Basic(t *testing.T) {
	content := `
version = 1.2.3
authors = Alice, Bob
emails  = alice@example.com, bob@example.com
homes   = https://example.com, https://bob.example.com
description = A cool library
created = 2024-06-15
`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.2.3" {
		t.Errorf("Version = %q, want 1.2.3", info.Version)
	}
	if len(info.Authors) != 2 || info.Authors[0] != "Alice" {
		t.Errorf("Authors = %v", info.Authors)
	}
	if len(info.Emails) != 2 || info.Emails[1] != "bob@example.com" {
		t.Errorf("Emails = %v", info.Emails)
	}
	if len(info.Homes) != 2 {
		t.Errorf("Homes = %v", info.Homes)
	}
	if info.Description != "A cool library" {
		t.Errorf("Description = %q", info.Description)
	}
	if info.Created.IsZero() {
		t.Errorf("Created should not be zero for %q", info.CreatedRaw)
	}
	if info.Created.Year() != 2024 {
		t.Errorf("Created year = %d, want 2024", info.Created.Year())
	}
}

func TestParseContent_WhitespaceForms(t *testing.T) {
	cases := []string{
		"version=1.0.0",
		"version =1.0.0",
		"version= 1.0.0",
		"version = 1.0.0",
		"  version   =   1.0.0  ",
	}
	for _, c := range cases {
		info, err := goversion.ParseContent(c, "")
		if err != nil {
			t.Errorf("ParseContent(%q) error: %v", c, err)
			continue
		}
		if info.Version != "1.0.0" {
			t.Errorf("ParseContent(%q) Version = %q, want 1.0.0", c, info.Version)
		}
	}
}

func TestParseContent_CaseInsensitiveKeys(t *testing.T) {
	content := `
VERSION = 2.0.0
AUTHORS = Dev One, Dev Two
EMAILS = dev1@x.com
HOMES = https://x.com
DESCRIPTION = test
`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "2.0.0" {
		t.Errorf("Version = %q", info.Version)
	}
	if len(info.Authors) != 2 {
		t.Errorf("Authors = %v", info.Authors)
	}
}

func TestParseContent_CommentSkipping(t *testing.T) {
	content := `
# this is a comment
// also a comment
version = 3.3.3
`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "3.3.3" {
		t.Errorf("Version = %q", info.Version)
	}
}

func TestParseContent_InlineComment(t *testing.T) {
	content := `version = 4.0.0 # stable`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "4.0.0" {
		t.Errorf("Version = %q (inline comment not stripped)", info.Version)
	}
}

func TestParseContent_ExtraFields(t *testing.T) {
	content := `
version = 1.0.0
license = MIT
build = 42
`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatal(err)
	}
	if info.Extra["license"] != "MIT" {
		t.Errorf("Extra[license] = %q", info.Extra["license"])
	}
	if info.Extra["build"] != "42" {
		t.Errorf("Extra[build] = %q", info.Extra["build"])
	}
}

func TestParseContent_SingleAuthor(t *testing.T) {
	content := `author = Solo Dev`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Authors) != 1 || info.Authors[0] != "Solo Dev" {
		t.Errorf("Authors = %v", info.Authors)
	}
}

// ---------------------------------------------------------------------------
// Date / time parsing
// ---------------------------------------------------------------------------

func TestParseContent_DateFormats(t *testing.T) {
	cases := []struct {
		input string
		year  int
	}{
		{"2024-01-15", 2024},
		{"2024-01-15 10:30:00", 2024},
		{"2024-01-15T10:30:00", 2024},
		{"2024-01-15T10:30:00Z", 2024},
		{"15/01/2024", 2024},
	}
	for _, tc := range cases {
		content := "created = " + tc.input
		info, err := goversion.ParseContent(content, "")
		if err != nil {
			t.Errorf("created=%q error: %v", tc.input, err)
			continue
		}
		if info.Created.IsZero() {
			t.Errorf("created=%q: got zero time", tc.input)
			continue
		}
		if info.Created.Year() != tc.year {
			t.Errorf("created=%q: year=%d want %d", tc.input, info.Created.Year(), tc.year)
		}
	}
}

func TestParseContent_UnixTimestamp(t *testing.T) {
	ts := time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC).Unix()
	content := "created = 1678838400"
	_ = ts
	info, _ := goversion.ParseContent(content, "")
	if info.Created.IsZero() {
		t.Error("expected non-zero time for unix timestamp")
	}
}

func TestParseContent_UnparsedDate(t *testing.T) {
	content := `created = not-a-date`
	info, err := goversion.ParseContent(content, "")
	if err != nil {
		t.Fatal(err)
	}
	if info.CreatedRaw != "not-a-date" {
		t.Errorf("CreatedRaw = %q", info.CreatedRaw)
	}
	if !info.Created.IsZero() {
		t.Errorf("Created should be zero for bad input")
	}
}

// ---------------------------------------------------------------------------
// ParseFile
// ---------------------------------------------------------------------------

func TestParseFile_AllCandidateNames(t *testing.T) {
	dir := t.TempDir()
	content := "version = 9.9.9\n"

	// Test a representative subset of the well-known names.
	names := []string{
		"VERSION",
		"VERSIONS",
		"VER",
		"__version__",
		"__VERSION__",
		"__version__.py",
		"VERSION.py",
		"VERSION.txt",
		"version.txt",
		"version.cfg",
		"version.ini",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path := writeFile(t, dir, name, content)
			info, err := goversion.ParseFile(path)
			if err != nil {
				t.Fatalf("ParseFile(%q): %v", name, err)
			}
			if info.Version != "9.9.9" {
				t.Errorf("Version = %q", info.Version)
			}
			// Clean up so the file doesn't interfere with FindIn tests.
			os.Remove(path)
		})
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := goversion.ParseFile("/this/path/does/not/exist/VERSION")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// Finder / FindIn
// ---------------------------------------------------------------------------

func TestFinder_FindIn(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "VERSION", "version = 0.1.2\nauthors = Hadi Cahyadi\n")

	f := goversion.NewFinder()
	info, err := f.FindIn(dir)
	if err != nil {
		t.Fatalf("FindIn: %v", err)
	}
	if info.Version != "0.1.2" {
		t.Errorf("Version = %q", info.Version)
	}
	if len(info.Authors) == 0 || info.Authors[0] != "Hadi Cahyadi" {
		t.Errorf("Authors = %v", info.Authors)
	}
}

func TestFinder_FindIn_NoFile(t *testing.T) {
	dir := t.TempDir()
	f := goversion.NewFinder()
	_, err := f.FindIn(dir)
	if err == nil {
		t.Error("expected error when no version file present")
	}
}

func TestFinder_WithFilename(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "my_custom_ver.txt", "version = 5.6.7\n")

	f := goversion.NewFinder(goversion.WithFilename("my_custom_ver.txt"))
	info, err := f.FindIn(dir)
	if err != nil {
		t.Fatalf("FindIn with custom filename: %v", err)
	}
	if info.Version != "5.6.7" {
		t.Errorf("Version = %q", info.Version)
	}
}

func TestFinder_WithExtraDirs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "VERSION", "version = 1.1.1\n")

	f := goversion.NewFinder(goversion.WithExtraDirs(dir))
	info, err := f.Find()
	if err != nil {
		t.Fatalf("Find with extra dir: %v", err)
	}
	if info.Version != "1.1.1" {
		t.Errorf("Version = %q", info.Version)
	}
}

func TestFinder_WithBaseName(t *testing.T) {
	f := goversion.NewFinder(goversion.WithBaseName("myapp"))
	// Just ensure no panic; actual directory search depends on environment.
	_ = f
}

// ---------------------------------------------------------------------------
// Convenience functions
// ---------------------------------------------------------------------------

func TestGetVersionFrom(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "VERSION", "version=7.8.9\n")
	v, err := goversion.GetVersionFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if v != "7.8.9" {
		t.Errorf("GetVersionFrom = %q", v)
	}
}

func TestGet_ExtraDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "VERSION", "version=3.2.1\n")
	info, err := goversion.Get(goversion.WithExtraDirs(dir))
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "3.2.1" {
		t.Errorf("Get Version = %q", info.Version)
	}
}

func TestGetVersion_ExtraDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "VERSION", "version=10.0.0\n")
	v, err := goversion.GetVersion(goversion.WithExtraDirs(dir))
	if err != nil {
		t.Fatal(err)
	}
	if v != "10.0.0" {
		t.Errorf("GetVersion = %q", v)
	}
}

// ---------------------------------------------------------------------------
// Info.String()
// ---------------------------------------------------------------------------

func TestInfo_String(t *testing.T) {
	info := &goversion.Info{
		Version:     "1.0.0",
		Authors:     []string{"Alice"},
		Emails:      []string{"alice@example.com"},
		Homes:       []string{"https://example.com"},
		Description: "test",
		FilePath:    "/tmp/VERSION",
		Extra:       map[string]string{},
	}
	s := info.String()
	if !strings.Contains(s, "1.0.0") {
		t.Errorf("String() does not contain version: %s", s)
	}
	if !strings.Contains(s, "Alice") {
		t.Errorf("String() does not contain author: %s", s)
	}
}

// ---------------------------------------------------------------------------
// Platform-specific path smoke tests
// ---------------------------------------------------------------------------

func TestSearchDirs_NoPanic(t *testing.T) {
	// Just ensure the search-dir builders don't panic on the current platform.
	f := goversion.NewFinder(goversion.WithBaseName("testapp"))
	// Call Find(); it will fail to find a file but must not panic.
	_, _ = f.Find()
}

func TestPlatform(t *testing.T) {
	t.Logf("running on GOOS=%s", runtime.GOOS)
}
