//go:build !mattn && !modernc

package store

import (
	_ "github.com/ncruces/go-sqlite3/driver"
)

const DriverName = "sqlite3"
