// Package goversion provides utilities to discover and parse version metadata
// from conventional version files (VERSION, __version__.py, VERSION.txt, etc.).
//
// It supports flexible key=value parsing with optional whitespace, multi-value
// fields (authors, emails, homes), date/time parsing for the "created" field,
// and a platform-aware search strategy that mirrors the lookup order used by
// well-known configuration libraries on Windows, Linux, and macOS.
//
// # Module Information
//
//   - Name:    go-version
//   - Author:  Hadi Cahyadi <cumulus13@gmail.com>
//   - Home:    https://github.com/cumulus13/go-version
package goversion

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// Info holds all metadata parsed from a version file.
type Info struct {
	// Version is the semver-like string, e.g. "1.2.3".
	Version string

	// Authors is a list of author names split on commas.
	Authors []string

	// Emails is a list of e-mail addresses split on commas.
	Emails []string

	// Homes is a list of project home URLs split on commas.
	Homes []string

	// Description is a free-form description string.
	Description string

	// Created holds the parsed creation date when the value is recognisable
	// as a date/time string. It is zero when parsing fails.
	Created time.Time

	// CreatedRaw is the original unparsed string for the "created" field.
	CreatedRaw string

	// FilePath is the absolute path of the file that was successfully read.
	FilePath string

	// Extra holds any key=value pairs that are not one of the well-known fields.
	Extra map[string]string
}

// String returns a human-readable summary of the Info.
func (i Info) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Version:     %s\n", i.Version)
	fmt.Fprintf(&sb, "Authors:     %s\n", strings.Join(i.Authors, ", "))
	fmt.Fprintf(&sb, "Emails:      %s\n", strings.Join(i.Emails, ", "))
	fmt.Fprintf(&sb, "Homes:       %s\n", strings.Join(i.Homes, ", "))
	fmt.Fprintf(&sb, "Description: %s\n", i.Description)
	if !i.Created.IsZero() {
		fmt.Fprintf(&sb, "Created:     %s\n", i.Created.Format(time.RFC3339))
	} else if i.CreatedRaw != "" {
		fmt.Fprintf(&sb, "Created:     %s (unparsed)\n", i.CreatedRaw)
	}
	fmt.Fprintf(&sb, "File:        %s\n", i.FilePath)
	return sb.String()
}

// ---------------------------------------------------------------------------
// Well-known version file names (case-insensitive match is done at runtime)
// ---------------------------------------------------------------------------

// versionFileNames is the ordered list of candidate filenames tried for every
// search directory.  The list is intentionally broad to cover common
// conventions across Python, Node, Rust, plain-text, and custom projects.
var versionFileNames = []string{
	"VERSION",
	"VERSIONS",
	"VER",
	"VERS",
	"__VERSION__",
	"__version__",
	"__version__.py",
	"VERSION.py",
	"VERSION.txt",
	"VERSION.md",
	"VERSION.cfg",
	"VERSION.ini",
	"VERSION.json",
	"VERSION.toml",
	"VERSION.yaml",
	"VERSION.yml",
	"VERSION.conf",
	"VERSION.config",
	"VERSION.properties",
	"VERSION.env",
	"VERSION.rc",
	"VERS.txt",
	"VER.txt",
	"ver.txt",
	"version.txt",
	"version.py",
	"version.cfg",
	"version.ini",
	"version.json",
	"version.toml",
	"version.yaml",
	"version.yml",
	"version.conf",
	"version.properties",
	"version.env",
	"version.rc",
	"version",
	"versions",
}

// ---------------------------------------------------------------------------
// Date/time layouts tried when parsing the "created" field
// ---------------------------------------------------------------------------

var dateLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04",
	"2006-01-02",
	"02/01/2006 15:04:05",
	"02/01/2006",
	"01/02/2006",
	"Jan 2, 2006",
	"January 2, 2006",
	"2 Jan 2006",
	"2 January 2006",
	"Mon, 02 Jan 2006 15:04:05 -0700",
	time.RFC1123Z,
	time.RFC1123,
	time.RFC822Z,
	time.RFC822,
	time.ANSIC,
	time.UnixDate,
}

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

// Option is a functional option for configuring a Finder.
type Option func(*Finder)

// WithBaseName overrides the base name used when building the search paths.
// By default the executable name (without extension) is used.
func WithBaseName(name string) Option {
	return func(f *Finder) { f.baseName = name }
}

// WithExtraDirs adds extra directories that are prepended to the search list.
func WithExtraDirs(dirs ...string) Option {
	return func(f *Finder) { f.extraDirs = append(f.extraDirs, dirs...) }
}

// WithFilename restricts the search to a single filename instead of the full
// candidate list.
func WithFilename(name string) Option {
	return func(f *Finder) { f.fixedFilename = name }
}

// ---------------------------------------------------------------------------
// Finder
// ---------------------------------------------------------------------------

// Finder discovers and parses version files.
type Finder struct {
	baseName      string
	extraDirs     []string
	fixedFilename string
}

// NewFinder creates a new Finder with the supplied options applied.
func NewFinder(opts ...Option) *Finder {
	f := &Finder{}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Find searches for a version file and returns the parsed Info.
// It returns an error when no version file can be located.
func (f *Finder) Find() (*Info, error) {
	dirs := f.searchDirs()
	candidates := f.candidateNames()

	for _, dir := range dirs {
		for _, name := range candidates {
			full := filepath.Join(dir, name)
			info, err := ParseFile(full)
			if err == nil {
				return info, nil
			}
		}
	}
	return nil, fmt.Errorf("go-version: no version file found (searched %d directories)", len(dirs))
}

// FindIn searches only inside dir (non-recursive).
func (f *Finder) FindIn(dir string) (*Info, error) {
	candidates := f.candidateNames()
	for _, name := range candidates {
		full := filepath.Join(dir, name)
		info, err := ParseFile(full)
		if err == nil {
			return info, nil
		}
	}
	return nil, fmt.Errorf("go-version: no version file found in %q", dir)
}

// candidateNames returns the list of filenames to try.
func (f *Finder) candidateNames() []string {
	if f.fixedFilename != "" {
		return []string{f.fixedFilename}
	}
	return versionFileNames
}

// searchDirs returns the platform-specific ordered list of directories to
// search, from most-specific to least-specific.
func (f *Finder) searchDirs() []string {
	var dirs []string

	// User-supplied extras come first.
	dirs = append(dirs, f.extraDirs...)

	switch runtime.GOOS {
	case "windows":
		dirs = append(dirs, windowsSearchDirs(f.baseName)...)
	default:
		dirs = append(dirs, unixSearchDirs(f.baseName)...)
	}

	return dedupe(dirs)
}

// ---------------------------------------------------------------------------
// Platform search-path builders
// ---------------------------------------------------------------------------

func windowsSearchDirs(baseName string) []string {
	var dirs []string

	home, _ := os.UserHomeDir()
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")
	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = `C:\ProgramData`
	}

	execDir := executableDir()
	cwd, _ := os.Getwd()

	addVariants := func(base, name string) {
		if base == "" || name == "" {
			return
		}
		dirs = append(dirs,
			filepath.Join(base, name),
		)
	}

	// 1. %USERPROFILE%\<name>
	if home != "" && baseName != "" {
		dirs = append(dirs, filepath.Join(home, baseName))
	}
	// 2. %APPDATA%\<name>
	if appData != "" {
		addVariants(appData, baseName)
	}
	// 3. %APPDATA%\Roaming\<name>  (APPDATA already points here on most systems,
	//    but we add the explicit sub-path for completeness)
	if appData != "" {
		dirs = append(dirs, filepath.Join(appData, "Roaming", baseName))
	}
	// 4. %LOCALAPPDATA%\<name>
	if localAppData != "" {
		addVariants(localAppData, baseName)
	}
	// 5. %PROGRAMDATA%\<name>
	if baseName != "" {
		dirs = append(dirs, filepath.Join(programData, baseName))
	}
	// 6. Executable directory
	if execDir != "" {
		dirs = append(dirs, execDir)
	}
	// 7. Current working directory
	if cwd != "" {
		dirs = append(dirs, cwd)
	}

	return dirs
}

func unixSearchDirs(baseName string) []string {
	var dirs []string

	home, _ := os.UserHomeDir()
	execDir := executableDir()
	cwd, _ := os.Getwd()

	addIf := func(d string) {
		if d != "" {
			dirs = append(dirs, d)
		}
	}

	// 1. ~/.<name>  (hidden config dir)
	if home != "" && baseName != "" {
		addIf(filepath.Join(home, "."+baseName))
	}
	// 2. ~/.config/<name>
	if home != "" && baseName != "" {
		addIf(filepath.Join(home, ".config", baseName))
	}
	// 3. $XDG_CONFIG_HOME/<name>
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" && baseName != "" {
		addIf(filepath.Join(xdg, baseName))
	}
	// 4. /etc/<name>
	if baseName != "" {
		addIf(filepath.Join("/etc", baseName))
	}
	// 5. /usr/local/etc/<name>
	if baseName != "" {
		addIf(filepath.Join("/usr/local/etc", baseName))
	}
	// 6. Executable directory
	addIf(execDir)
	// 7. Current working directory
	addIf(cwd)

	return dirs
}

// executableDir returns the directory containing the running executable.
func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	return filepath.Dir(resolved)
}

// dedupe removes duplicate entries while preserving order.
func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

// kvRe matches lines of the form:  key = value  (spaces around '=' are optional)
var kvRe = regexp.MustCompile(`(?i)^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)

// ParseFile reads path and parses it as a version file.
// It returns an error when the file cannot be opened.
func ParseFile(path string) (*Info, error) {
	f, err := os.Open(path) // #nosec G304 — intentional file read
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := &Info{
		FilePath: path,
		Extra:    make(map[string]string),
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip blank lines and comment lines (# or //).
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		m := kvRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(m[1]))
		val := strings.TrimSpace(m[2])
		// Strip inline comments.
		if idx := strings.Index(val, " #"); idx != -1 {
			val = strings.TrimSpace(val[:idx])
		}

		switch key {
		case "version":
			info.Version = val
		case "author", "authors":
			info.Authors = splitList(val)
		case "email", "emails":
			info.Emails = splitList(val)
		case "home", "homes", "url", "urls":
			info.Homes = splitList(val)
		case "description", "desc":
			info.Description = val
		case "created", "create_date", "createdate", "date":
			info.CreatedRaw = val
			info.Created = parseTime(val)
		default:
			info.Extra[key] = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("go-version: scanning %q: %w", path, err)
	}

	return info, nil
}

// ParseContent parses version metadata from an in-memory string.
// filePath is stored in the returned Info for reference; it may be empty.
func ParseContent(content, filePath string) (*Info, error) {
	info := &Info{
		FilePath: filePath,
		Extra:    make(map[string]string),
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		m := kvRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(m[1]))
		val := strings.TrimSpace(m[2])
		if idx := strings.Index(val, " #"); idx != -1 {
			val = strings.TrimSpace(val[:idx])
		}

		switch key {
		case "version":
			info.Version = val
		case "author", "authors":
			info.Authors = splitList(val)
		case "email", "emails":
			info.Emails = splitList(val)
		case "home", "homes", "url", "urls":
			info.Homes = splitList(val)
		case "description", "desc":
			info.Description = val
		case "created", "create_date", "createdate", "date":
			info.CreatedRaw = val
			info.Created = parseTime(val)
		default:
			info.Extra[key] = val
		}
	}

	return info, nil
}

// ---------------------------------------------------------------------------
// Package-level convenience functions
// ---------------------------------------------------------------------------

// Get is the zero-configuration entry point.  It creates a default Finder and
// calls Find(), deriving the base name from the running executable.
func Get(opts ...Option) (*Info, error) {
	return NewFinder(opts...).Find()
}

// GetVersion returns only the version string.
func GetVersion(opts ...Option) (string, error) {
	info, err := Get(opts...)
	if err != nil {
		return "", err
	}
	return info.Version, nil
}

// GetVersionFrom reads the version from a specific file.
func GetVersionFrom(path string) (string, error) {
	info, err := ParseFile(path)
	if err != nil {
		return "", err
	}
	return info.Version, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// splitList splits a comma-separated string and trims each element.
func splitList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// parseTime tries every known layout until one succeeds.
func parseTime(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	// Try Unix timestamp (integer seconds).
	var ts int64
	if _, err := fmt.Sscanf(s, "%d", &ts); err == nil && ts > 0 {
		return time.Unix(ts, 0).UTC()
	}
	return time.Time{}
}
