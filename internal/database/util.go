package database

import "database/sql"

// CheckSupportDatabaseType check support Database driver.
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
