package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/boltdb/bolt"
)

const (
	// StorageTypeBoltDB константа для типа хранилища boltdb.
	StorageTypeBoltDB          = "boltdb"
	migrateDataFilePathDefault = "data/migrations.db"
)

// BoltDB тип хранилища boltdb.
type BoltDB struct {
	connect *bolt.DB
}

// Init инициализация соединения с хранилищем.
func (b *BoltDB) Init(cfg *Config) error {
	if cfg.Path == "" {
		cfg.Path = migrateDataFilePathDefault
	}

	var err error
	b.connect, err = bolt.Open(cfg.Path, 0600, nil)
	if err != nil {
		return err
	}

	return nil
}

// PreInit подготовка файлов и прочего к работе.
func (b *BoltDB) PreInit(cfg *Config) error {
	if cfg.Path == "" {
		cfg.Path = migrateDataFilePathDefault
	}

	dir := filepath.Dir(cfg.Path)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}

	return nil
}

// Close закрытие соединения с хранилищем.
func (b *BoltDB) Close() error {
	if b.connect != nil {
		return b.connect.Close()
	}

	return nil
}

// CreateProjectDB создание под новый проект/базу новый бакет.
func (b *BoltDB) CreateProjectDB(projectName, dbName string) error {
	if b.connect == nil {
		return errors.New("connect is lost")
	}

	err := b.connect.Update(func(tx *bolt.Tx) error {
		bp, createProjectBucketERR := tx.CreateBucketIfNotExists([]byte(projectName))
		if createProjectBucketERR != nil {
			return createProjectBucketERR
		}

		// внутри бакета с именем проекта созадаем бакет с именем базы данных.
		if _, createDBBucketERR := bp.CreateBucketIfNotExists([]byte(dbName)); createDBBucketERR != nil {
			return createDBBucketERR
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// CheckMigration проверить выполнялась ли миграция.
func (b *BoltDB) CheckMigration(projectName, dbName, version string) (bool, error) {
	if b.connect == nil {
		return false, errors.New("connect is lost")
	}

	found := false
	err := b.connect.View(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(projectName))
		if bp == nil {
			return errors.New("Project " + projectName + " not exists")
		}
		bkt := bp.Bucket([]byte(dbName))
		if bkt == nil {
			return errors.New("Database " + dbName + " not exists")
		}

		if v := bkt.Get([]byte(version)); v != nil {
			mi := Migrate{}
			err := json.Unmarshal(v, &mi)
			if err != nil {
				return err
			}
			found = true
		}

		return nil
	})

	return found, err
}

// Up выполнить миграцию.
func (b *BoltDB) Up(post *Migrate) error {
	if b.connect == nil {
		return errors.New("connect is lost")
	}

	return b.connect.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(post.Project))
		if bp == nil {
			return errors.New("Project " + post.Project + " not exists")
		}
		bkt := bp.Bucket([]byte(post.Database))
		if bkt == nil {
			return errors.New("Database " + post.Database + " not exists")
		}
		encoded, err := json.Marshal(post)
		if err != nil {
			return err
		}

		return bkt.Put([]byte(post.Version), encoded)
	})
}

// GetLast получить список последних миграций.
func (b *BoltDB) GetLast(projectName, dbName string, skipNoRollback bool, limit *int) ([]Migrate, error) {
	if b.connect == nil {
		return nil, errors.New("connect is lost")
	}

	migrates := []Migrate{}
	err := b.connect.Update(func(tx *bolt.Tx) error {
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

			// добавим только откатываемые миграции или все, если флаг (только откатываемые миграции) == false
			if mi.RollFlag || !skipNoRollback {
				migrates = append(migrates, mi)
			}

			return nil
		})

		return err
	})

	result := []Migrate{}

	if err == nil {
		// отсортируем в правильном порядке
		sort.Slice(migrates, func(i, j int) bool {
			return migrates[i].Version > migrates[j].Version
		})

		// Проверим что есть требуемое кол-во миграций
		if limit != nil && len(migrates) >= *limit {
			// возьмём с начала count миграций
			result = migrates[:*limit]
		} else {
			result = migrates
		}
	}

	return result, err
}

// Delete откатить миграцию.
func (b *BoltDB) Delete(post *Migrate) error {
	if b.connect == nil {
		return errors.New("connect is lost")
	}

	err := b.connect.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(post.Project))
		if bp == nil {
			return errors.New("Project " + post.Project + " not exists")
		}
		bkt := bp.Bucket([]byte(post.Database))
		if bkt == nil {
			return errors.New("Database " + post.Database + " not exists")
		}

		// TODO: сделать проверку вдруг есть миграции после текущей

		return bkt.Delete([]byte(post.Version))
	})

	return err
}
