package backend

import "embed"

// Migrations contains the immutable SQL migration history used by cmd/migrate.
//
//go:embed migrations/*.sql
var Migrations embed.FS
