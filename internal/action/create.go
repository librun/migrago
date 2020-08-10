package action

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

func MakeCreate(cfgPath, name, mode, project, db string) error {
	// get config
	cfg := yamlConfig{}
	configFile, err := os.Open(cfgPath)

	if err != nil {
		return fmt.Errorf("config open error: %s", err)
	}

	defer configFile.Close()
	configByte, err := ioutil.ReadAll(configFile)

	if err != nil {
		return fmt.Errorf("config read error: %s", err)
	}
	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return fmt.Errorf("config format error: %s", err)
	}

	// get migration directory
	directory := ""

	for projectName, value := range cfg.Projects {
		if projectName != project {
			continue
		}
		for _, migrations := range value.Migrations {
			ok := false

			if directory, ok = migrations[db]; !ok {
				continue
			} else {
				break
			}
		}
	}
	if directory == "" {
		return errors.New("invalid project or db")
	}
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err = os.MkdirAll(directory, 0777); err != nil {
			return errors.New(fmt.Sprintf("create directory %s error: %s", directory, err))
		}
	}

	// create file
	if mode == "both" {
		for _, modeItem := range []string{"up", "down"} {
			if err = createFile(name, modeItem, directory); err != nil {
				return err
			}
		}
		return nil
	}

	return createFile(name, mode, directory)
}

func createFile(name, mode, directory string) error {
	now := time.Now()
	filename := fmt.Sprintf(
		"%d%02d%02d_%02d%02d_%s_%s.sql",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		name,
		mode,
	)
	_, err := os.Create(fmt.Sprintf("%s/%s", directory, filename))

	return err
}
