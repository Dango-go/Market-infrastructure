// Package authdb embeds the auth service SQL migrations so they ship inside the binary and
// can be applied at startup without a separate migration tool.
package authdb

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrations returns the filesystem rooted at the migrations directory, containing the
// ordered *.up.sql / *.down.sql files.
func Migrations() fs.FS {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		panic(err) // embedded path is constant; a failure here is a build-time bug
	}
	return sub
}
