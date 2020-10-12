package database

import "database/sql"

// CheckSupportDatabaseType checks support database driver.
func CheckSupportDatabaseType(dbType string) bool {
	support := false

	for _, dbDriver := range sql.Drivers() {
		if dbDriver == dbType {
			support = true
			break
		}
	}

	return support
}
