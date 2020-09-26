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

## Использование
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


# Конфигурация

Мигратор migrago использует файл конфигурации в формате `yaml`.

### Пример файла конфигурации
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

## Блоки файла конфигурации

#### Блок хранения миграций `migration_storage`

|Аттрибут|Пример|Обязательный|Описание|
|--------|------|------------|--------|
|**storage_type**|postgres|да|Тип БД для хранения миграций (поддерживаемые типы: postgres, boltdb)|
|**dsn**|postgres://postgres:postgres@localhost:5432/migrago?sslmode=disable|для sql|Только для типа БД `postgres` Реквизиты для подключения к БД|
|**schema**|public|для postgres|Только для типа БД `postgres` схема для подключения|
|**path**|data/migrations.db|нет|Только для типа БД `boltdb` путь хранения файла с миграциями|

#### Блок баз данных `databases`
Необходимо указать уникальное имя БД

|Аттрибут|Пример|Обязательный|Описание|
|--------|------|------------|--------|
|**type**|postgres|да|Тип БД (поддерживаемые типы: postgres, mysql, clickhouse, sqlite3)|
|**dsn**|postgres://docker:docker@localhost/postgres?sslmode=disable|да|Реквизиты для подключения к БД|
|**schema**|test|нет|Только для типа БД `postgres` схема для подключения|

#### Блок проектов `projects`
Необходимо указать уникальное имя проекта

|Аттрибут|Описание|
|--------|------|
|**migrations**|Массив Имя БД и путь до миграций|

# Описание cli

#### Глобальные опции

|Опция|Алиас|Пример|Обязательная|Описание|
|-----|-----|------|------------|--------|
|config|-c --config| -c sample.yaml|да|Путь до конфига|

#### Иниализация работы мигратора `init`

Создаёт требуемое окружение (для *postgres* создаст таблицу для миграции, для *boltdb* директорию если её не было)

#### Обновить структуры до последней версии `up`
##### Опции
|Опция|Алиас|Пример|Обязательная|Описание|
|-----|-----|------|------------|--------|
|project|-p --project|-p project1|нет|Применить миграции только определённого проекта|
|database|-d --db --database|-d postgres1|нет|Применить миграции только определённой БД|

##### Пример
```bash
./migrago -c config.yaml up -p project1 -d postgres1
```

#### Откат миграций `down`
##### Опции
|Опция|Пример|Обязательная|Описание|
|-----|------|------------|--------|
|project|project1|да|имя проекта|
|db|postgres1|да|имя БД|
|len|1|да|количество откатываемых миграций|
|no-skip||нет|не пропускать не откатываемые миграции|

##### Пример
```bash
./migrago -c config.yaml down -p project1 -d postgres1 -l 1
```

#### Просмотр применённых миграций `list`
##### Опции
|Опция|Пример|Обязательная|Описание|
|-----|------|------------|--------|
|project|project1|да|имя проекта|
|db|postgres1|да|имя БД|
|len|1|нет|количество выводимых миграций|
|no-skip||нет|не пропускать не откатываемые миграции|

##### Пример
```bash
./migrago -c config.yaml list -p project1 -d postgres1
```

# Требования к файлам миграции
При указании новой миграции необходимо создать файлы:  
`%временная метка%`_`%имя миграции%`_up.sql и  
`%временная метка%`_`%имя миграции%`_down.sql (Если требуется создать откатываемую миграцию)  
Временная метка в формате: _ГГГГММДДЧЧММСС_  

#### Пример:
Файл _20190307010200_book_up.sql_
```sql
CREATE TABLE book
(
  id                      SERIAL,
  title                   VARCHAR(255)
  body                    TEXT,
  PRIMARY KEY(id)
);
```

Файл отката миграции _20190307010200_book_down.sql_
```sql
DROP TABLE book;
```
