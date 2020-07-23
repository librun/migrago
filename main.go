package main

import (
	"errors"
	"log"
	"os"

	"github.com/ne-ray/migrago/internal/action"
	"github.com/ne-ray/migrago/internal/storage"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "migrago"
	app.Version = "1.1.1"
	app.Usage = "cli-migration"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: `config, c`, Usage: `path to configuration file`, Required: true},
	}
	app.Commands = []cli.Command{
		{
			Name:        "up",
			Usage:       "Upgrade a database to its latest structure",
			Description: "To upgrade a database to its latest structure, you should apply all available new migrations using this command",
			ArgsUsage:   "",
			Flags: []cli.Flag{
				cli.StringFlag{Name: `project, p`, Usage: `project name`},
				cli.StringFlag{Name: `database, db, d`, Usage: `database name`},
			},
			Action: func(c *cli.Context) error {
				mStorage, err := storage.Init(c.GlobalString("config"))
				if err != nil {
					return err
				}
				defer func() {
					if err := mStorage.Close(); err != nil {
						log.Println(err)
					}
				}()

				var project *string
				var database *string
				if c.IsSet("project") {
					p := c.String("project")
					project = &p
				}
				if c.IsSet("database") {
					d := c.String("database")
					database = &d
				}

				if err := action.MakeUp(mStorage, c.GlobalString("config"), project, database); err != nil {
					return err
				}

				log.Println("Migration up is successfully")

				return nil
			},
		},
		{
			Name:        "down",
			Usage:       "Revert (undo) one or multiple migrations",
			Description: "To revert (undo) one or multiple migrations that have been applied before, you can run this command",
			ArgsUsage:   "",
			Flags: []cli.Flag{
				cli.StringFlag{Name: `project, p`, Usage: `project name`, Required: true},
				cli.StringFlag{Name: `database, db, d`, Usage: `database name`, Required: true},
				cli.IntFlag{Name: `limit, l`, Usage: `limit revert migrations`, Required: true, Value: 1},
				cli.BoolFlag{Name: `no-skip`, Usage: `not skip migration with rollback is false`},
			},
			Action: func(c *cli.Context) error {
				mStorage, err := storage.Init(c.GlobalString("config"))
				if err != nil {
					return err
				}
				defer func() {
					if err := mStorage.Close(); err != nil {
						log.Println(err)
					}
				}()

				rollbackCount := c.Int("limit")
				if rollbackCount < 1 {
					return errors.New("limit revert migration is not define")
				}
				// флаг пропускать неоткатываемые миграции
				skip := true
				if c.IsSet("no-skip") {
					skip = false
				}

				if err := action.MakeDown(mStorage, c.GlobalString("config"), c.String("project"), c.String("database"), rollbackCount, skip); err != nil {
					return err
				}

				log.Println("Rollback is successfully")

				return nil
			},
		},
		{
			Name:        "list",
			Usage:       "show list migrations",
			Description: "Show list migrations that have been applied before, you can run this command",
			ArgsUsage:   "",
			Flags: []cli.Flag{
				cli.StringFlag{Name: `project, p`, Usage: `project name`, Required: true},
				cli.StringFlag{Name: `database, db, d`, Usage: `database name`, Required: true},
				cli.IntFlag{Name: `limit, l`, Usage: `limit revert migrations`},
				cli.BoolFlag{Name: `no-skip`, Usage: `not skip migration with rollback is false`},
			},
			Action: func(c *cli.Context) error {
				mStorage, err := storage.Init(c.GlobalString("config"))
				if err != nil {
					return err
				}
				defer func() {
					if err := mStorage.Close(); err != nil {
						log.Println(err)
					}
				}()

				var rollbackCount *int
				if c.IsSet("limit") {
					if limit := c.Int("limit"); limit > 0 {
						rollbackCount = &limit
					} else {
						log.Fatalln("limit revert migration is not correct")
					}
				}
				// флаг пропускать неоткатываемые миграции
				skip := true
				if c.IsSet("no-skip") {
					skip = false
				}

				if err := action.MakeList(mStorage, c.GlobalString("config"), c.String("project"), c.String("database"), rollbackCount, skip); err != nil {
					log.Fatalln(err)
				}

				return nil
			},
		},
		{
			Name:        "init",
			Usage:       "Initialize storage",
			Description: "Initialize storage (for example boltdb - create dir, sql - create table with migrations)",
			ArgsUsage:   "",
			Action: func(c *cli.Context) error {
				if err := storage.PreInit(c.GlobalString("config")); err != nil {
					return err
				}

				log.Println("init storage is successfully")

				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
