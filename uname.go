package goupd

// those tables contain mapping from uname to GOOS / GOARCH values

// Keys may contain wildcards (* or ?) meant to be matched using https://pkg.go.dev/path#Match
// Those are also suitable for bash case matching

// uname -s
var UnameOsMatch = map[string]string{
	"Linux":    "linux",
	"CYGWIN_*": "windows", // CYGWIN_NT-5.1
	"MINGW*":   "windows",
	"WIN32":    "windows",
	"WINNT":    "windows",
	"Windows":  "windows",
	"Darwin*":  "darwin",
	"FreeBSD":  "freebsd",
}

// uname -m
var UnameArchMatch = map[string]string{
	"x86_64":     "amd64",
	"amd64":      "amd64",
	"i686":       "386",
	"powerpc":    "ppc",
	"ppc":        "ppc",
	"ppc64":      "ppc64",
	"armv7l":     "arm",
	"armv6l":     "arm",
	"armv7b":     "armbe",
	"armv6b":     "armbe",
	"aarch64":    "arm64",
	"aarch64_be": "arm64be",
	"armv8b":     "arm64be",
	"armv8l":     "arm64",
	"mips":       "mips",
}
