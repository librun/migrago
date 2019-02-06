package internal

import (
	"github.com/ne-ray/migrago.git/internal/model"
)

//DbBolt instance BoltDB for save migration data
var DbBolt *model.Bolt

//MigratePostfixUp postfix for file up migrate
const MigratePostfixUp = "_up.sql"

//MigratePostfixDown postfix for file up migrate
const MigratePostfixDown = "_down.sql"

const migrateDataFilePathDefault = "data/migrations.db"

//InitInstance init instance for work migrations
func InitInstance(pathDataFile string) error {
	if pathDataFile == "" {
		pathDataFile = migrateDataFilePathDefault
	}

	dbBolt, err := model.InitBoltDb(pathDataFile)
	if err != nil {
		return err
	}
	DbBolt = dbBolt

	return nil
}
