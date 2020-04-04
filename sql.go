package migrator

func insertMigrationVersion() string {
	return "INSERT INTO dbmigrator_versioning (migrator_version, migrator_timestamp,file_path) VALUES ($1,$2,$3)"
}

func selectMigrationVersion() string {
	return "SELECT id,migrator_version,migrator_timestamp from dbmigrator_versioning order by id desc LIMIT 1"
}

func createMigrationTableIfNotExists() string {
	return `CREATE TABLE IF NOT EXISTS dbmigrator_versioning
	(
		id SERIAL NOT NULL PRIMARY KEY,
		migrator_version INT NOT NULL,
		migrator_timestamp bigint NOT NULL,
		file_path VARCHAR(255) NOT NULL
	);`
}
