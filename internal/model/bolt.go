package model

import (
	"errors"
	"strconv"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/ne-ray/migrago.git/internal/helper"
)

type (
	Migrate struct {
		MigrationID	uint64
		FileName	string
		RollFlag	bool
	}
	Bolt struct {
		Connect	*bolt.DB
	}
)

func InitBoltDb(path string) (*Bolt, error) {
	boltInstance := &Bolt{}
	dbBolt, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return boltInstance, err
	}
	boltInstance.Connect = dbBolt

	return boltInstance, nil
}

func (b *Bolt) CreateProjectDB(projectName, dbName string) error {
	err := b.Connect.Update(func(tx *bolt.Tx) error {
		bp, createProjectBucketERR := tx.CreateBucketIfNotExists([]byte(projectName))
		if createProjectBucketERR != nil {
			return createProjectBucketERR
		}
		//внутри бакета с именем проекта созадаем бакет с именем базы данных
		if _, createDBBucketERR  := bp.CreateBucketIfNotExists([]byte(dbName)); createDBBucketERR != nil {
			return createDBBucketERR
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (b *Bolt) CheckMigration(projectName, dbName, migrateFileName string) (bool, error) {
	found := false
	err := b.Connect.View(func(tx *bolt.Tx) error {
		errorFoundMigration := errors.New("Migration is already exists")
		bp := tx.Bucket([]byte(projectName))
		if bp == nil {
			return errors.New("Project " + projectName + " not exists")
		}
		bkt := bp.Bucket([]byte(dbName))
		if bkt == nil {
			return errors.New("Database " + dbName + " not exists")
		}

		err := bkt.ForEach(func(k, v []byte) error {
			mi := Migrate{}
			err := json.Unmarshal(v, &mi)
			if err != nil {
				return err
			}
			if mi.FileName == migrateFileName {
				//если текущий файл уже выполнялся, то больше его не запускаем
				found = true

				//Кинем ошибку, что бы завершить итератор
				return errorFoundMigration
			}

			return nil
		})

		if err != nil && err != errorFoundMigration {
			return err
		}

		return nil
	})

	return found, err
}

func (b *Bolt) Up(projectName, dbName string, post *Migrate) error {
	return b.Connect.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(projectName))
		if bp == nil {
			return errors.New("Project " + projectName + " not exists")
		}
		bkt := bp.Bucket([]byte(dbName))
		if bkt == nil {
			return errors.New("Database " + dbName + " not exists")
		}
		//инкреминтируем запись в бакете
		migrationID, err := bkt.NextSequence()
		if err != nil {
			return err
		}
		post.MigrationID = migrationID
		encoded, err := json.Marshal(post)
		if err != nil {
			return err
		}

		return bkt.Put(helper.Itob(post.MigrationID), encoded)
	})
}

func (b *Bolt) GetLast(projectName, dbName string, count int) ([]Migrate, error) {
	migrates := map[uint64]Migrate{}
	maxMigrate := uint64(0)
	err := b.Connect.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(projectName))
		if bp == nil {
			return errors.New("Project " + projectName + " not exists")
		}
		bkt := bp.Bucket([]byte(dbName))
		if bkt == nil {
			return errors.New("Database " + dbName + " not exists")
		}
		err := bkt.ForEach(func(k, v []byte) error {
			mi := Migrate{}
			err := json.Unmarshal(v, &mi)
			if err != nil {
				return err
			}
			migrates[mi.MigrationID] = mi
			if maxMigrate < mi.MigrationID {
				maxMigrate = mi.MigrationID
			}

			return nil
		})

		return err
	})

	result := []Migrate{}

	if err == nil {
		//Проверим что есть требуемое кол-во миграций
		if len(migrates) >= count {
			for i := 0; len(result) < count; i++ {
				if migrate, ok := migrates[maxMigrate - uint64(i)]; ok {
					result = append(result, migrate)
				}
			}
		} else {
			err = errors.New("Not have " + strconv.Itoa(count) + " migration")
		}
	}

	return result, err
}

func (b *Bolt) Delete(projectName, dbName string, ID uint64) ( error) {
	err := b.Connect.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(projectName))
		if bp == nil {
			return errors.New("Project " + projectName + " not exists")
		}
		bkt := bp.Bucket([]byte(dbName))
		if bkt == nil {
			return errors.New("Database " + dbName + " not exists")
		}

		return bkt.Delete(helper.Itob(ID))
	})

	return err
}
