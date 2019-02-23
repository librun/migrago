package main

import (
	"log"
	"os"
	"strconv"

	"github.com/ne-ray/migrago/internal"
	"github.com/ne-ray/migrago/internal/action"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "migrago"
	app.Version = "0.1.0"
	app.Usage = "cli-migration"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: `config, c`, Usage: `path to configuration file`},
		cli.StringFlag{Name: `datafile, f`, Usage: `path to migrate data file`},
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
				if err := internal.InitInstance(c.GlobalString("datafile")); err != nil {
					log.Fatalln(err)
				}
				defer internal.DbBolt.Connect.Close()

				projects := []string{}
				databases := []string{}
				if c.IsSet("project") {
					projects = append(projects, c.String("project"))
				}
				if c.IsSet("database") {
					databases = append(databases, c.String("database"))
				}
				config, err := internal.InitConfig(c.GlobalString("config"), projects, databases)
				if err != nil {
					log.Fatalln(err)
				}

				if err := action.MakeUp(&config); err != nil {
					log.Fatalln(err)
				}

				log.Println("Migration up is successfully")

				return nil
			},
		},
		{
			Name:        "down",
			Usage:       "Revert (undo) one or multiple migrations",
			Description: "To revert (undo) one or multiple migrations that have been applied before, you can run this command",
			ArgsUsage:   "project db count",
			Action: func(c *cli.Context) error {
				if err := internal.InitInstance(c.String("datafile")); err != nil {
					log.Fatalln(err)
				}
				defer internal.DbBolt.Connect.Close()

				if len(c.Args()) < 3 {
					log.Fatalln("args not complete")
				}
				projectName := c.Args().Get(0)
				dbName := c.Args().Get(1)
				rollbackCount, err := strconv.Atoi(c.Args().Get(2))
				if err != nil {
					log.Fatalln(err)
				}

				config, err := internal.InitConfig(c.GlobalString("config"), []string{projectName}, []string{dbName})
				if err != nil {
					log.Fatalln(err)
				}

				project, err := config.GetProject(projectName)
				if err != nil {
					log.Fatalln(err)
				}

				if _, err := project.GetDB(dbName); err != nil {
					log.Fatalln(err)
				}

				if err := action.MakeDown(&project, dbName, rollbackCount); err != nil {
					log.Fatalln(err)
				}

				log.Println("Rollback is successfully")

				return nil
			},
		},
	}

	app.Run(os.Args)
}
