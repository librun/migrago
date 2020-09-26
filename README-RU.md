Migrago
=======

Migrago это инструмент для миграций SQL-like баз данных.

# Установка

## Установка из релизных бинарников

Скачайте бинарный файл из подготовленных [релизов](https://github.com/librun/migrago/releases/latest) и поместите его в 
директорию `$GOPATH/bin`.

### Linux

    wget -qO- "https://github.com/librun/migrago/releases/download/v1.1.0/migrago-0.1.0-amd64_linux.tar.gz" \
        | tar -zOx "migrago-0.6.0-amd64_linux/migrago" > "$GOPATH"/bin/migrago && chmod +x "$GOPATH"/bin/migrago

## Установка из исходников

    go get https://github.com/librun/migrago@v1.0.0

# Использование
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

## Глобальные опции

Для работы migrago требуется файл конфигурации. 

    migrago -c pat/to/config.yaml command 

### Пример файла конфигурации

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

Блок хранения миграций.

|Аттрибут|Пример|Обязательный|Описание|
|--------|------|------------|--------|
|**storage_type**|postgres|да|Тип БД для хранения миграций (поддерживаемые типы: postgres, boltdb)|
|**dsn**|postgres://postgres:postgres@localhost:5432/migrago?sslmode=disable|для sql|Только для типа БД `postgres` Реквизиты для подключения к БД|
|**schema**|public|для postgres|Только для типа БД `postgres` схема для подключения|
|**path**|data/migrations.db|нет|Только для типа БД `boltdb` путь хранения файла с миграциями|

### databases

Блок баз данных. Необходимо указывать уникальные имена для баз данных.

|Аттрибут|Пример|Обязательный|Описание|
|--------|------|------------|--------|
|**type**|postgres|да|Тип БД (поддерживаемые типы: postgres, mysql, clickhouse, sqlite3)|
|**dsn**|postgres://docker:docker@localhost/postgres?sslmode=disable|да|Реквизиты для подключения к БД|
|**schema**|test|нет|Только для типа БД `postgres` схема для подключения|

### projects

Блок проектов. Необходимо указывать уникальные имена для проектов.

|Аттрибут|Описание|
|--------|------|
|**migrations**|Массив Имя БД и путь до миграций|

## Команды

### init

Команда init создаёт требуемое окружение для дальнейшей работы migrago. Для *postgres* Будет создана таблица `migration`, 
в базе данных, которая указана в блоке `migration_storage` файла конфигурации. Для *boltdb* будет содана директория для 
файла базы данных если её не было.

    $ migrago -c config.yaml init
    2020/09/26 16:17:38 init storage is successfully

### up

Применение миграций. Может использоваться без дополнительных опций. В этом случае будут выполненыы 
все доступные миграции всех доступных проектов. 

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

Можно дополнительно указать проект и базу данных, для которых необходимо выполнить миграции:    

    $ migrago -c config.yaml up -p testproject -d postgres
    2020/09/26 17:02:46 Project: testproject
    2020/09/26 17:02:46 ----------
    2020/09/26 17:02:46 DB: postgres
    2020/09/26 17:02:46 migration success: 20200427_170000_create_table_test
    2020/09/26 17:02:46 migration success: 20200925_150000_update_table_test
    2020/09/26 17:02:46 Completed migrations: 2 of 2
    2020/09/26 17:02:46 ----------
    2020/09/26 17:02:46 Migration up is successfully

|Опция|Алиас|Пример|Обязательная|Описание|
|-----|-----|------|------------|--------|
|project|-p --project|-p testproject|нет|Применить миграции только определённого проекта|
|database|-d --db --database|-d postgres|нет|Применить миграции только определённой БД|

### down

Откат миграций. Необходимо указать проект, базу данных и количество миграций для отката. Опции `project`, `db` и `len` 
обязательны. Указанное количество откатываемых мыграций должно быть меньше, либо быть равным количеству существующих 
миграций для указанных проекта и базы данных. 

    $ migrago -c config.yaml down -p testproject -d postgres1 -l 1
    2020/09/26 16:15:10 migration: 20200427_170000_create_table_test roolback completed
    2020/09/26 16:15:10 migration: 20200925_150000_update_table_test roolback completed
    2020/09/26 16:15:10 Rollback is successfully

|Опция|Пример|Обязательная|Описание|
|-----|------|------------|--------|
|project|testproject|да|имя проекта|
|db|postgres|да|имя БД|
|len|1|да|количество откатываемых миграций|
|no-skip||нет|не пропускать не откатываемые миграции|

### list

Просмотр применённых миграций. Опции `project` и `db` обязательны. 

    $ migrago -c config.yaml list -p testproject -d postgres
    2020/09/26 16:55:56 List migrations:
    2020/09/26 16:55:56 migration: 20200427_170000_create_table_test
    2020/09/26 16:55:56 migration: 20200925_150000_update_table_test

|Опция|Пример|Обязательная|Описание|
|-----|------|------------|--------|
|project|project1|да|имя проекта|
|db|postgres1|да|имя БД|
|len|1|нет|количество выводимых миграций|
|no-skip||нет|не пропускать не откатываемые миграции|

### create #WIP 

Создание новой SQL миграции. Опции `project` и `db` обязательны.

    $ migrago -c config.yaml create -p dashboard -d postgres

|Опция|Пример|Обязательная|Описание|
|-----|------|------------|--------|
|project, p|project1|да|имя проекта|
|db, d|postgres1|да|имя БД|
|name, n|create_table_user|нет|имя для миграции|
|mode, m|up|нет|тип создаваемой миграции up/down/both (default: up)|

# Требования к файлам миграции

При указании новой миграции необходимо создать файлы:  
`%временная метка%`_`%имя миграции%`_up.sql и  
`%временная метка%`_`%имя миграции%`_down.sql (Если требуется создать откатываемую миграцию)  
Временная метка в формате: _ГГГГММДДЧЧММСС_  

## Пример:

### Файл _20190307010200_book_up.sql_

```sql
CREATE TABLE book
(
  id                      SERIAL,
  title                   VARCHAR(255)
  body                    TEXT,
  PRIMARY KEY(id)
);
```

### Файл отката миграции _20190307010200_book_down.sql_

```sql
DROP TABLE book;
```
