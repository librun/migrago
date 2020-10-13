package storage

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Allowed database storage types.
const (
	TypeBoltDB   = "boltdb"
	TypePostgres = "postgres"
)

type (
	// Storage describes methods for working with a migration storage.
	Storage interface {
		PreInit(cfg *Config) error
		Init(cfg *Config) error
		Close() error
		CreateProjectDB(projectName, dbName string) error
		CheckMigration(projectName, dbName, version string) (bool, error)
		Up(post *Migrate) error
		GetLast(projectName, dbName string, skipNoRollback bool, limit *int) ([]Migrate, error)
		Delete(post *Migrate) error
	}

	// Config contains storage credentials information.
	Config struct {
		StorageType string `yaml:"storage_type"`
		Path        string `yaml:"path"`
		DSN         string `yaml:"dsn"`
		Schema      string `yaml:"schema"`
	}

	configFull struct {
		MigrationStorage Config `yaml:"migration_storage"`
	}

	// Migrate is the model for table migration.
	Migrate struct {
		Project   string
		Database  string
		Version   string
		ApplyTime int64
		RollFlag  bool
	}
)

// New creates instance for work with migrations.
func New(pathConfigFile string) (Storage, error) {
	cfg, err := parseConfig(pathConfigFile)
	if err != nil {
		return nil, err
	}

	s := getStorage(cfg.StorageType)

	if err := s.Init(cfg); err != nil {
		return nil, err
	}

	return s, nil
}

// PreInit runs preinit function.
func PreInit(pathConfigFile string) error {
	cfg, err := parseConfig(pathConfigFile)
	if err != nil {
		return err
	}

	s := getStorage(cfg.StorageType)

	if err := s.PreInit(cfg); err != nil {
		return err
	}

	if err := s.Close(); err != nil {
		return err
	}

	return nil
}

// parseConfig gets and returns the part of the config associated
// with the migrator storage.
func parseConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config open error: %s", err)
	}
	defer configFile.Close()

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("config read error: %s", err)
	}

	cfg := configFull{}

	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return nil, fmt.Errorf("config format error: %s", err)
	}

	return &cfg.MigrationStorage, nil
}

func getStorage(typeS string) Storage {
	var s Storage

	switch typeS {
	case TypeBoltDB:
		s = &BoltDB{}
	case TypePostgres:
		s = &PostgreSQL{}
	default: // use BoltDB by default for backward compatibility
		s = &BoltDB{}
	}

	return s
}
