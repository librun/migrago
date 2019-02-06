# Мигратор SQL-like Базданных

## Описание конфига `yaml`

#### Блок базданных `databases`
Необходимо указать уникальное имя БД

|Аттрибут|Пример|Описание|
|--------|--------|------|
|**type**|postgres|Тип БД (поддерживаемые типы: postgres, mysql, clickhouse, sqlite3)|
|**dsn**|postgres://docker:docker@localhost/postgres?sslmode=disable|Реквизиты для подключения к БД|
|**schema**|test|_Необязательный аттрибут_ Только для типа БД `postgres` схема для подключения|

#### Блок проектов `projects`
Необходимо указать уникальное имя проекта

|Аттрибут|Пример|Описание|
|--------|--------|------|
|**migrations**| |Массив Имя БД и путь до миграций|

#### Пример конфига
```yaml
projects:
  you-project:
    migrations:
    - migrate: dir/for/migrations
  
databases:
  migrate:
    type: postgres
    dsn: "postgres://docker:docker@localhost/postgres?sslmode=disable"
    schema: "test"
  stats_clickhouse:
    type: clickhouse
    dsn: "clickhouse://docker:docker@localhost/clickhouse?sslmode=disable"
```

## Описание cli
#### Глобальные опции

|Опция|Алиас|Пример|Описание|
|-----|-----|------|--------|
|config|-c --config| -c sample.yaml| Путь до конфига|

#### Обновить структуры до последней версии `up`
##### Опции
|Опция|Алиас|Пример|Описание|
|-----|-----|------|--------|
|project|-p --project|-p you-project|_Необязательная опция_ Применить миграции только определённого проекта|
|database|-d --db --database|-d migrate|_Необязательная опция_ Применить миграции только определённой БД|

##### Пример
```bash
./migrate -c config.yaml up -d migrate -p you-project
```

#### Откат миграций `down`
##### Аргументы
|Опция|Пример|Описание|
|-----|------|--------|
|project|you-project|_Обязательный аргумент_ имя проекта|
|db|migrate|_Обязательный аргумент_ имя БД|
|count|1|_Обязательный аргумент_ количество откатываемых миграций|

##### Пример
```bash
./migrate -c config.yaml down you-project migrate 1
```

## Требования к файлам миграции
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
