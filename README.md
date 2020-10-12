# Migrago
[![golangci-lint](https://github.com/librun/migrago/workflows/golangci-lint/badge.svg)](https://github.com/librun/migrago/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/librun/migrago)](https://goreportcard.com/report/github.com/librun/migrago)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/librun/migrago)
[![GitHub All Releases](https://img.shields.io/github/downloads/librun/migrago/total)](https://github.com/librun/migrago/releases)

Cli Migrator for SQL-like Databases.

[readme in Russian](/README-RU.md)

## Install
### Installation from release binaries
Download the binary from the prepared [releases](https://github.com/librun/migrago/releases/latest) and put it to
directory `$GOPATH/bin`.

#### Linux
    wget -qO- "https://github.com/librun/migrago/releases/download/v1.1.0/migrago-1.1.0-amd64_linux.tar.gz" \
        | tar -zOx "migrago-1.1.0-amd64_linux/migrago" > "$GOPATH"/bin/migrago && chmod +x "$GOPATH"/bin/migrago

### Installing from source
    go get https://github.com/librun/migrago@v1.1.0

## Usage
```text
USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   up       Upgrade a database to its latest structure
   down     Revert (undo) one or multiple migrations
   list     show list migrations
   init     Initialize storage
   create   create new migration
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  path to configuration file
   --help, -h                show help
   --version, -v             print the version
```

Migrago requires a configuration file to work. 

    migrago -c pat/to/config.yaml command
    
### Sample config file
<details>
<summary>config-example.yaml</summary>

```yaml
migration_storage:
  storage_type: "postgres" # "boltdb"
  dsn: "postgres://postgres:postgres@localhost:5432/migrago?sslmode=disable"
  schema: "public"
  path: "data/migrations.db"

projects:
  project1:
    migrations:
    - postgres1: dir/for/migrations_postgres
    - clickhouse1: dir/for/migrations_clickhouse
  
databases:
  postgres1:
    type: postgres
    dsn: "postgres://postgres:postgres@localhost/database?sslmode=disable"
    schema: "test"
  clickhouse1:
    type: clickhouse
    dsn: "tcp://host1:9000?username=user&password=qwerty&database=clicks"
```
</details>

### migration_storage
Migration storage unit.

|Attribute|Required|Description|
|--------|------------|--------|
|**storage_type**|yes|Database type for storing migrations (supported types: postgres, boltdb)|
|**dsn**|yes for sql|For DB type `postgres` only. Requisites for connecting to the DB|
|**schema**|yes for postgres|Only for DB type `postgres` schema for connection|
|**path**|yes for boltdb|For DB type `boltdb` only. Path to store the file with migrations|

### projects
Projects unit. You must provide unique names for projects. Typically, one configuration file uses only one project, but 
it is possible to specify several projects. For each project you can specify the paths for the migration files for each 
used databases in the project.

### databases
Database unit. You must provide unique names for the databases. Contains configuration for connecting to databases
which are used in projects.

## Commands
### init
The init command creates the required environment for migrago. For *postgres* will be created table `migration`,
in the database that is specified in the `migration_storage` block of the configuration file. For *boltdb* a directory 
will be created for the database file if did not exist.

    $ migrago -c config.yaml init
    2020/09/26 16:17:38 init storage is successfully

### up
Applying migrations. It can be used without additional options. In this case all available migrations will be applied of 
all available projects.

    $ migrago -c config.yaml up
    2020/09/26 16:44:15 Project: testproject
    2020/09/26 16:44:15 ----------
    2020/09/26 16:44:15 DB: postgres
    2020/09/26 16:44:15 migration success: 20200427_170000_create_table_test
    2020/09/26 16:44:15 migration success: 20200925_150000_update_table_test
    2020/09/26 16:44:15 Completed migrations: 2 of 2
    2020/09/26 16:44:15 ----------
    2020/09/26 16:44:15 DB: clickhouse
    2020/09/26 16:44:15 migration success: 20200123_200800_create_table_test
    2020/09/26 16:44:15 Completed migrations: 1 of 1
    2020/09/26 16:44:15 ----------
    2020/09/26 16:44:15 Migration up is successfully

You can additionally specify the project and database for which you want apply the migration:    

    $ migrago -c config.yaml up -p testproject -d postgres
    2020/09/26 17:02:46 Project: testproject
    2020/09/26 17:02:46 ----------
    2020/09/26 17:02:46 DB: postgres
    2020/09/26 17:02:46 migration success: 20200427_170000_create_table_test
    2020/09/26 17:02:46 migration success: 20200925_150000_update_table_test
    2020/09/26 17:02:46 Completed migrations: 2 of 2
    2020/09/26 17:02:46 ----------
    2020/09/26 17:02:46 Migration up is successfully

|Option|Alias|Required|Description|
|-----|-----|------------|--------|
|project|-p --project|no|Apply migrations to only a specific project|
|database|-d --db --database|no|Apply migrations only to a specific database|

### down
Rolling back migrations. You must specify the project, database, and number of migrations to rollback. The `project`, `db` 
and `len` options are required. The specified number of rolled back migrations must be less or equal to the number of 
existing migrations for the specified project and database.

    $ migrago -c config.yaml down -p testproject -d postgres1 -l 1
    2020/09/26 16:15:10 migration: 20200427_170000_create_table_test roolback completed
    2020/09/26 16:15:10 migration: 20200925_150000_update_table_test roolback completed
    2020/09/26 16:15:10 Rollback is successfully

|Option|Required|Description|
|-----|------------|--------|
|project|yes|Project name|
|db|yes|Database name|
|len|yes|Number of rolled back migrations|
|no-skip|no|Do not skip non-rollback migrations|

### list
View applied migrations. The `project` and `db` options are required.

    $ migrago -c config.yaml list -p testproject -d postgres
    2020/09/26 16:55:56 Migrations list:
    2020/09/26 16:55:56 migration: 20200427_170000_create_table_test
    2020/09/26 16:55:56 migration: 20200925_150000_update_table_test

|Option|Required|Description|
|-----|------------|--------|
|project|yes|Project name|
|db|yes|Database name|
|len|no|Number of migrations to output|
|no-skip|no|Do not skip non-rollback migrations|

### create
Creating a new SQL migration. The options `project`, `db` and `name` are required.

    $ migrago -c config.yaml create -p testproject -d postgres -n create_table_test
    2020/09/27 05:41:40 migration: 20200927_054140_create_table_test_up.sql created
    2020/09/27 05:41:40 migration: 20200927_054140_create_table_test_down.sql created
    2020/09/27 05:41:40 Migration successfully created

|Option|Required|Description|
|-----|------------|--------|
|project, p|yes|Project name|
|db, d|yes|Database name|
|name, n|yes|Name for migration|
|mode, m|no|Type of migration to create up/down/both (default: up)|

# Migration file requirements
When specifying a new migration, you need to create files:  
`%time%_%name%_up.sql` and  
`%time%_%name%_down.sql` (If you want to create a rollback migration)  
Time format: `YYYYMMDD_HHMMCC`

## Example:
### File `20190307_010200_book_up.sql`
```sql
CREATE TABLE book
(
  id                      SERIAL,
  title                   VARCHAR(255)
  body                    TEXT,
  PRIMARY KEY(id)
);
```

### Migration rollback file `2019030_7010200_book_down.sql`
```sql
DROP TABLE book;
```
