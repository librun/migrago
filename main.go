package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/librun/migrago/internal/action"
	"github.com/librun/migrago/internal/storage"
	"github.com/urfave/cli"
)

// Version displays service version in semantic versioning (http://semver.org/).
// Can be replaced while compiling with flag "-ldflags "-X main.Version=${VERSION}"".
var Version = "develop"

func main() {
	app := cli.NewApp()
	app.Name = "migrago"
	app.Version = Version
	app.Usage = "cli-migration"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "config, c", Usage: "Path to configuration file", Required: true},
	}
	app.Commands = []cli.Command{
		getCommandUp(),
		getCommandDown(),
		getCommandList(),
		getCommandInit(),
		getCommandCreate(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func getCommandUp() cli.Command {
	return cli.Command{
		Name:        "up",
		Usage:       "Upgrade a database to its latest structure",
		Description: "To upgrade a database to its latest structure, you should apply all available new migrations using this command",
		ArgsUsage:   "",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "project, p", Usage: "Project name"},
			cli.StringFlag{Name: "database, db, d", Usage: "Database name"},
		},
		Action: func(c *cli.Context) error {
			mStorage, err := storage.New(c.GlobalString("config"))
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
	}
}

func getCommandDown() cli.Command {
	return cli.Command{
		Name:        "down",
		Usage:       "Revert (undo) one or multiple migrations",
		Description: "To revert (undo) one or multiple migrations that have been applied before, you can run this command",
		ArgsUsage:   "",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "project, p", Usage: "Project name", Required: true},
			cli.StringFlag{Name: "database, db, d", Usage: "Database name", Required: true},
			cli.IntFlag{Name: "limit, l", Usage: "Limit revert migrations", Required: true, Value: 1},
			cli.BoolFlag{Name: "no-skip", Usage: "Not skip migration with rollback is false"},
		},
		Action: func(c *cli.Context) error {
			mStorage, err := storage.New(c.GlobalString("config"))
			if err != nil {
				return err
			}
			defer func() {
				if err := mStorage.Close(); err != nil {
					log.Println(err)
				}
			}()

			project := c.String("project")
			if project == "" {
				return errors.New("project required")
			}

			db := c.String("db")
			if db == "" {
				return errors.New("database required")
			}

			rollbackCount := c.Int("limit")
			if rollbackCount < 1 {
				return errors.New("limit revert migration is not define")
			}

			// Flag for skip non-rolling migrations.
			skip := true
			if c.IsSet("no-skip") {
				skip = false
			}

			if err := action.MakeDown(mStorage, c.GlobalString("config"), project, db, rollbackCount, skip); err != nil {
				return fmt.Errorf("down: %w", err)
			}

			log.Println("Rollback is successfully")

			return nil
		},
	}
}

func getCommandList() cli.Command {
	return cli.Command{
		Name:        "list",
		Usage:       "Show migrations list",
		Description: "Show migrations that have been applied",
		ArgsUsage:   "",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "project, p", Usage: "Project name", Required: true},
			cli.StringFlag{Name: "database, db, d", Usage: "Database name", Required: true},
			cli.IntFlag{Name: "limit, l", Usage: "Limit revert migrations"},
			cli.BoolFlag{Name: "no-skip", Usage: "Not skip migration with rollback is false"},
		},
		Action: func(c *cli.Context) error {
			mStorage, err := storage.New(c.GlobalString("config"))
			if err != nil {
				return err
			}
			defer func() {
				if err := mStorage.Close(); err != nil {
					log.Println(err)
				}
			}()

			project := c.String("project")
			if project == "" {
				return errors.New("project required")
			}

			db := c.String("db")
			if db == "" {
				return errors.New("database required")
			}

			var rollbackCount *int
			if c.IsSet("limit") {
				if limit := c.Int("limit"); limit > 0 {
					rollbackCount = &limit
				} else {
					log.Fatalln("limit revert migration is not correct")
				}
			}

			// Flag for skip non-rolling migrations.
			skip := true
			if c.IsSet("no-skip") {
				skip = false
			}

			if err := action.MakeList(mStorage, c.GlobalString("config"), project, db, rollbackCount, skip); err != nil {
				log.Fatalln(err)
			}

			return nil
		},
	}
}

func getCommandInit() cli.Command {
	return cli.Command{
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
	}
}

func getCommandCreate() cli.Command {
	return cli.Command{
		Name:        "create",
		Usage:       "Create new migration",
		Description: "Create new empty migration file in project directory",
		ArgsUsage:   "",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "project, p", Usage: "Project name", Required: true},
			cli.StringFlag{Name: "database, db, d", Usage: "Database name", Required: false},
			cli.StringFlag{Name: "name, n", Usage: "File name", Required: false},
			cli.StringFlag{Name: "mode, m", Usage: "Migration file type [up|down|both] (default: up)", Required: false},
		},
		Action: func(c *cli.Context) error {
			project := c.String("project")
			if project == "" {
				return errors.New("project required")
			}

			db := c.String("db")
			if db == "" {
				return errors.New("database required")
			}

			name := c.String("name")
			if name == "" {
				return errors.New("migration name required")
			}

			mode := c.String("mode")
			if mode == "" {
				mode = action.CreateModeBoth
			}

			if mode != action.CreateModeUp && mode != action.CreateModeDown && mode != action.CreateModeBoth {
				return fmt.Errorf("invalid mode: %s", mode)
			}

			if err := action.MakeCreate(c.GlobalString("config"), name, mode, project, db); err != nil {
				log.Fatalln(err)
			}

			log.Println("Migration successfully created")

			return nil
		},
	}
}
