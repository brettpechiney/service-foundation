// +build mage

package main

import "github.com/magefile/mage/sh"

var Default = Migrate

// Migrate starts the database migration process.
func Migrate() error {
	return sh.RunV("sql-migrate", "up")
}
