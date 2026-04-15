# go-version

[![Go Reference](https://pkg.go.dev/badge/github.com/cumulus13/go-version.svg)](https://pkg.go.dev/github.com/cumulus13/go-version)
[![Go Report Card](https://goreportcard.com/badge/github.com/cumulus13/go-version)](https://goreportcard.com/report/github.com/cumulus13/go-version)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**go-version** discovers and parses version metadata from conventional version
files (`VERSION`, `__version__.py`, `VERSION.txt`, …) using a platform-aware
search strategy.

| Field | Example value | Notes |
|---|---|---|
| `version` | `1.2.3` | Any semver-like string |
| `authors` | `Alice, Bob` | Comma-separated list |
| `emails` | `a@x.com, b@x.com` | Comma-separated list |
| `homes` | `https://…, https://…` | Comma-separated list |
| `description` | `A cool library` | Free text |
| `created` | `2024-06-01` | Parsed to `time.Time` when possible |
| _any other key_ | | Stored in `Info.Extra` |

---

## Installation

```bash
go get github.com/cumulus13/go-version
```

## Quick start

```go
package main

import (
    "fmt"
    goversion "github.com/cumulus13/go-version"
)

func main() {
    info, err := goversion.Get()
    if err != nil {
        panic(err)
    }
    fmt.Println(info)
}
```

---

## Version file format

Keys are **case-insensitive**. Whitespace around `=` is optional.
Lines starting with `#` or `//` are treated as comments.

```ini
# my project version file
version     = 1.2.3
authors     = Hadi Cahyadi, Jane Doe
emails      = cumulus13@gmail.com, jane@example.com
homes       = https://github.com/cumulus13/go-version
description = A version-file discovery library for Go
created     = 2024-06-01
license     = MIT          # extra fields are stored in Info.Extra
```

All of the following `=` styles are valid:

```
version=1.0.0
version =1.0.0
version= 1.0.0
version = 1.0.0
```

---

## Supported file names

The library tries every name in the table below for each search directory.

| Filename | Filename | Filename |
|---|---|---|
| `VERSION` | `VERSIONS` | `VER` |
| `VERS` | `__VERSION__` | `__version__` |
| `__version__.py` | `VERSION.py` | `VERSION.txt` |
| `VERSION.md` | `VERSION.cfg` | `VERSION.ini` |
| `VERSION.json` | `VERSION.toml` | `VERSION.yaml` |
| `VERSION.yml` | `VERSION.conf` | `VERSION.config` |
| `VERSION.properties` | `VERSION.env` | `VERSION.rc` |
| `version.txt` | `version.py` | `version.cfg` |
| `version.ini` | `version.json` | `version.toml` |
| `version.yaml` | `version.yml` | `version.conf` |
| `version.properties`| `version.env` | `version.rc` |
| `version` | `versions` | `ver.txt` |

---

## Search order

### Windows

1. `%USERPROFILE%\<appname>`
2. `%APPDATA%\<appname>`
3. `%APPDATA%\Roaming\<appname>`
4. `%LOCALAPPDATA%\<appname>`
5. `%PROGRAMDATA%\<appname>`
6. Executable directory
7. Current working directory

### Linux / macOS

1. `~/.<appname>`
2. `~/.config/<appname>`
3. `$XDG_CONFIG_HOME/<appname>`
4. `/etc/<appname>`
5. `/usr/local/etc/<appname>`
6. Executable directory
7. Current working directory

---

## API reference

### Zero-config

```go
// Search automatically; uses the executable name as the app base name.
info, err := goversion.Get()

// Get only the version string.
v, err := goversion.GetVersion()

// Parse a specific file.
v, err := goversion.GetVersionFrom("/path/to/VERSION")
```

### Finder with options

```go
f := goversion.NewFinder(
    goversion.WithBaseName("myapp"),          // override app name for path building
    goversion.WithExtraDirs("/opt/myapp"),    // prepend extra search directories
    goversion.WithFilename("release.cfg"),    // search for this specific filename only
)

info, err := f.Find()          // search all configured directories
info, err := f.FindIn("/tmp")  // search a single directory
```

### In-memory parsing

```go
content := `version = 2.0.0\nauthors = Dev`
info, err := goversion.ParseContent(content, "")

// Or parse a file directly:
info, err := goversion.ParseFile("/path/to/VERSION")
```

### Info struct

```go
type Info struct {
    Version     string
    Authors     []string
    Emails      []string
    Homes       []string
    Description string
    Created     time.Time   // zero when unparseable
    CreatedRaw  string      // original string
    FilePath    string
    Extra       map[string]string
}
```

---

## Supported date formats for `created`

The parser tries the following formats (in order):

- `RFC3339` / `RFC3339Nano`
- `2006-01-02T15:04:05` / `2006-01-02 15:04:05`
- `2006-01-02T15:04` / `2006-01-02 15:04`
- `2006-01-02`
- `02/01/2006 15:04:05` / `02/01/2006`
- `01/02/2006`
- `Jan 2, 2006` / `January 2, 2006`
- `2 Jan 2006` / `2 January 2006`
- RFC1123, RFC1123Z, RFC822, RFC822Z, ANSIC, UnixDate
- Unix timestamp (integer seconds)

When none of the formats match, `Info.Created` is a zero `time.Time` and
`Info.CreatedRaw` contains the original string.

---

## License

MIT © 2024 Hadi Cahyadi &lt;cumulus13@gmail.com&gt;

---

## 👤 Author
        
[Hadi Cahyadi](mailto:cumulus13@gmail.com)
    

[![Buy Me a Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/cumulus13)

[![Donate via Ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/cumulus13)
 
[Support me on Patreon](https://www.patreon.com/cumulus13)
