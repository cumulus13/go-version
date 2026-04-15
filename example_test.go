package goversion_test

import (
	"fmt"
	"os"
	"path/filepath"

	goversion "github.com/cumulus13/go-version"
)

// ExampleGet demonstrates zero-configuration usage: the library searches
// platform-specific directories for a version file that belongs to the
// running executable.
func ExampleGet() {
	// Create a temporary directory with a VERSION file to make the example
	// deterministic in tests.
	dir, _ := os.MkdirTemp("", "goversion-example-*")
	defer os.RemoveAll(dir)

	_ = os.WriteFile(filepath.Join(dir, "VERSION"), []byte(`
version     = 1.2.3
authors     = Hadi Cahyadi, Jane Doe
emails      = cumulus13@gmail.com, jane@example.com
homes       = https://github.com/cumulus13/go-version
description = A version-file discovery library for Go
created     = 2024-06-01
`), 0o644)

	info, err := goversion.Get(goversion.WithExtraDirs(dir))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("Version:", info.Version)
	fmt.Println("Authors:", info.Authors)
	fmt.Println("Emails:", info.Emails)
	fmt.Println("Description:", info.Description)
	fmt.Println("File:", info.FilePath != "")

	// Output:
	// Version: 1.2.3
	// Authors: [Hadi Cahyadi Jane Doe]
	// Emails: [cumulus13@gmail.com jane@example.com]
	// Description: A version-file discovery library for Go
	// File: true
}

// ExampleGetVersionFrom shows how to read a version from a known file path.
func ExampleGetVersionFrom() {
	dir, _ := os.MkdirTemp("", "goversion-example-*")
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "__version__.py")
	_ = os.WriteFile(path, []byte("version = 0.9.1\n"), 0o644)

	v, _ := goversion.GetVersionFrom(path)
	fmt.Println(v)

	// Output:
	// 0.9.1
}

// ExampleParseContent shows in-memory parsing without touching the filesystem.
func ExampleParseContent() {
	content := `
version     = 2.0.0
author      = Dev One
email       = dev@example.com
description = In-memory parse example
`
	info, _ := goversion.ParseContent(content, "")
	fmt.Println(info.Version)
	fmt.Println(info.Authors[0])

	// Output:
	// 2.0.0
	// Dev One
}

// ExampleNewFinder_withFilename shows how to search for a non-standard filename.
func ExampleNewFinder_withFilename() {
	dir, _ := os.MkdirTemp("", "goversion-example-*")
	defer os.RemoveAll(dir)

	_ = os.WriteFile(filepath.Join(dir, "release.cfg"), []byte("version = 3.1.4\n"), 0o644)

	f := goversion.NewFinder(
		goversion.WithFilename("release.cfg"),
		goversion.WithExtraDirs(dir),
	)
	info, _ := f.Find()
	fmt.Println(info.Version)

	// Output:
	// 3.1.4
}
