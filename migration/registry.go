// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

var migrationFileRE = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

// GetAll loads migrations from embedded SQL files.
func GetAll() []Migration {
	migrations, err := loadMigrations()
	if err != nil {
		panic(err)
	}

	return migrations
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(sqlFiles, "sql")
	if err != nil {
		return nil, fmt.Errorf("read migration sql dir: %w", err)
	}

	type pair struct {
		version string
		slug    string
		up      string
		down    string
	}
	byVersion := map[string]*pair{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		match := migrationFileRE.FindStringSubmatch(name)
		if match == nil {
			return nil, fmt.Errorf("invalid migration filename: %s", name)
		}

		version, slug, direction := match[1], match[2], match[3]
		body, err := sqlFiles.ReadFile(path.Join("sql", name))
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}

		sql := strings.TrimSpace(string(body))
		if sql == "" {
			return nil, fmt.Errorf("empty migration file: %s", name)
		}

		p := byVersion[version]
		if p == nil {
			p = &pair{version: version, slug: slug}
			byVersion[version] = p
		}
		if p.slug != slug {
			return nil, fmt.Errorf("migration %s has conflicting slugs: %s and %s", version, p.slug, slug)
		}

		if direction == "up" {
			p.up = sql
		} else {
			p.down = sql
		}
	}

	versions := make([]string, 0, len(byVersion))
	for version := range byVersion {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	migrations := make([]Migration, 0, len(versions))
	for _, version := range versions {
		p := byVersion[version]
		if p.up == "" {
			return nil, fmt.Errorf("migration %s missing .up.sql", version)
		}
		if p.down == "" {
			return nil, fmt.Errorf("migration %s missing .down.sql", version)
		}

		up, down := p.up, p.down
		migrations = append(migrations, Migration{
			Version:     version,
			Description: descriptionFromSlug(p.slug),
			Up: func(db *sql.DB) error {
				return exec(db, up)
			},
			Down: func(db *sql.DB) error {
				return exec(db, down)
			},
		})
	}

	return migrations, nil
}

func descriptionFromSlug(slug string) string {
	parts := strings.Split(slug, "_")
	if len(parts) == 0 {
		return slug
	}

	parts[0] = strings.ToUpper(parts[0][:1]) + parts[0][1:]

	return strings.Join(parts, " ")
}
