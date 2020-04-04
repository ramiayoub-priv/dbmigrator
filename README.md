# dbmigrator
Golang db migration for postgres

## This was intended for a personal project as a shared library. Currently works only with Postgres.

## Usage

`import (migrator "github.com/ramiayoub-priv/dbmigrator")`

`migrator.CheckAndRunMigrations(db, "./migrations")`
Where db is *sql.DB

## What it does

1. Creates a table caled dbmigrator_versioning if it does not already exist
`CREATE TABLE IF NOT EXISTS dbmigrator_versioning
	(
		id SERIAL NOT NULL PRIMARY KEY,
		migrator_version INT NOT NULL,
		migrator_timestamp bigint NOT NULL,
		file_path VARCHAR(255) NOT NULL
	);`
  
2. Reads the latest version from dbmigrator_versioning (starting from 0)
3. Reads the `./migrations` directory
Files must be `something__<version>*, for example `somefile__1.sql, `somefile__2.sql`
The version must be an integer, and the double underscore __ is reserved and MUST appear directly before the number
4. Will compare the version to the existing one in the DB.
5. Will run the migrations against the DB, starting for existing version + 1, in ascending order.

If it is of use to you, feel free to modify, use, do whatever you want with it
