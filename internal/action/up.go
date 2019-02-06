package action

import (
	"log"
	"os"
	"sort"
	"strings"
	"io/ioutil"
	"github.com/ne-ray/migrago.git/internal"
	"github.com/ne-ray/migrago.git/internal/model"
	i "github.com/ne-ray/migrago.git/internal/initialize"
)

//MakeUp do migrate up
func MakeUp(config *internal.Config) (error) {
	for _, project := range config.Projects {
		log.Println("Project: " + project.Name)
		for _, migration := range project.Migrations {
			log.Println("DB: " + migration.Database.Name)
			//создаем бакет по имени проекта
			if err := internal.DbBolt.CreateProjectDB(project.Name, migration.Database.Name); err != nil {
				return err
			}

			//список всех файлов
			filesInDir, err := ioutil.ReadDir(migration.Path)
			if err != nil {
				return err
			}

			var keys []string
			for _, f := range filesInDir {
				fileName := f.Name()
				//получаем список файлов на создание миграций, если имя короче 8 символов пропускаем
				if len(fileName) > 7 && fileName[len(fileName)-7:] == internal.MigratePostfixUp {
					keys = append(keys, fileName[:len(fileName)-7])
				}
			}
			//сортируем список миграций по дате создания
			sort.Strings(keys)

			//откроем соединение с БД
			db, errDB := i.InitDB(migration.Database.Type, migration.Database.Dsn, migration.Database.Schema)
			if errDB != nil {
				return errDB
			}
			for _, fileName := range keys {
				if check, err := internal.DbBolt.CheckMigration(project.Name, migration.Database.Name, fileName); check {
					continue
				} else if err != nil {
					db.Close()
					return err
				}

				content, err := ioutil.ReadFile(migration.Path + fileName + internal.MigratePostfixUp)
				if err != nil {
					db.Close()
					return err
				}

				query := string(content)
				if strings.TrimSpace(query) != "" {
					//выполнение всех запросов из текущего файла
					if errExec := db.Exec(query); errExec != nil {
						db.Close()
						return errExec
					}
				}

				log.Println("migration: " + fileName + " complete")

				post := &model.Migrate{
					FileName: fileName,
					RollFlag: true,
				}

				//если файла с окончанием down.sql не существует, то указываем, что эта миграция не откатываемая
				if _, err := os.Stat(migration.Path + fileName + internal.MigratePostfixDown); os.IsNotExist(err) {
					post.RollFlag = false
				}

				if err := internal.DbBolt.Up(project.Name, migration.Database.Name, post); err != nil {
					db.Close()
					return err
				}
			}

			if err := db.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
