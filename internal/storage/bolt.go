package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/boltdb/bolt"
)

const migrateDataFilePathDefault = "data/migrations.db"

// BoltDB represents a collection of buckets persisted to a file on disk.
type BoltDB struct {
	connect *bolt.DB
}

// Init creates and opens a database at the given path.
// If the file does not exist then it will be created automatically.
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

// PreInit creates dir fo database file is not exists.
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

// Close releases all database resources.
// All transactions must be closed before closing the database.
func (b *BoltDB) Close() error {
	if b.connect != nil {
		return b.connect.Close()
	}

	return nil
}

// CreateProjectDB creates a new bucket for project and database if it doesn't
// already exist.
func (b *BoltDB) CreateProjectDB(projectName, dbName string) error {
	if b.connect == nil {
		return errors.New("connect is lost")
	}

	err := b.connect.Update(func(tx *bolt.Tx) error {
		bp, createProjectBucketERR := tx.CreateBucketIfNotExists([]byte(projectName))
		if createProjectBucketERR != nil {
			return createProjectBucketERR
		}

		// Create a bucket with the database name inside the bucket with the name of the project.
		if _, createDBBucketERR := bp.CreateBucketIfNotExists([]byte(dbName)); createDBBucketERR != nil {
			return createDBBucketERR
		}

		return nil
	})

	return err
}

// CheckMigration checks the migration was done successfully.
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

// Up runs migration up.
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

// GetLast gets a list of recent migrations.
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
			if err := json.Unmarshal(v, &mi); err != nil {
				return err
			}

			// Add rollback migrations only (or all if flag `rollback migrations only` == false).
			if mi.RollFlag || !skipNoRollback {
				migrates = append(migrates, mi)
			}

			return nil
		})

		return err
	})

	result := []Migrate{}

	if err == nil {
		// Sort in the correct order.
		sort.Slice(migrates, func(i, j int) bool {
			return migrates[i].Version > migrates[j].Version
		})

		// Check that there is the required number of migrations.
		if limit != nil && len(migrates) >= *limit {
			result = migrates[:*limit]
		} else {
			result = migrates
		}
	}

	return result, err
}

// Delete calls migration down.
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

		// TODO: check if there are migrations after the current one.

		return bkt.Delete([]byte(post.Version))
	})

	return err
}
