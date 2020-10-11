package storage

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type (
	// Storage интерфейс хранилища.
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

	// Config конфиг для хранилища.
	Config struct {
		StorageType string `yaml:"storage_type"`
		Path        string `yaml:"path"`
		DSN         string `yaml:"dsn"`
		Schema      string `yaml:"schema"`
	}

	configFull struct {
		MigrationStorage Config `yaml:"migration_storage"`
	}

	// Migrate модель миграции.
	Migrate struct {
		Project   string
		Database  string
		Version   string
		ApplyTime int64
		RollFlag  bool
	}
)

// Init init instance for work migrations.
func Init(pathConfigFile string) (Storage, error) {
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

// PreInit run preinit function.
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

// parseConfig получим часть конфига связанную с хранилищем мигратора.
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
	case StorageTypeBoltDB:
		s = &BoltDB{}
	case StorageTypePostgreSQL:
		s = &PostgreSQL{}
	default: // по умолчанию используем boltdb (для обратной совместимости)
		s = &BoltDB{}
	}

	return s
}
