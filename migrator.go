package migrator

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
)

type dbMigratorSchema struct {
	id        int
	version   int
	timestamp int64
}

// CheckAndRunMigrations Will check if there is dbmigrator_versioning table and create it if needed
// Will then run migrations in the migrationDir against the provided db
func CheckAndRunMigrations(db *sql.DB, migrationDir string) {

	err := createMigratorTableIfNotExists(db)
	if err != nil {
		log.Fatalf("Migration failed on createMigrationTableIfNotExists due to %v", err)
	}

	version, err := checkVersion(db)
	if err != nil {
		log.Fatalf("Migration failed on checkVersion due to %v", err)
	}

	log.Printf("Current schema version is %d", version)

	m, err := readMigrationDirectory(migrationDir)

	if err != nil {
		log.Fatalf("Migration failed on readMigrationDirectory due to %v", err)
	}

	err = applyMigrations(db, &m, version)

	if err != nil {
		log.Fatalf("Migration failed on applyMigrations due to %v", err)
	}

}

func createMigratorTableIfNotExists(db *sql.DB) error {

	sql := createMigrationTableIfNotExists()

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	log.Printf("Will exec %s \n", sql)
	_, err = tx.Exec(sql)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err

}

func checkVersion(db *sql.DB) (int, error) {
	sql := selectMigrationVersion()
	stmt, err := db.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		log.Println("Error preparing statement", err)
		return -1, err
	}

	rows, err := stmt.Query()
	defer rows.Close()
	if err != nil {
		log.Println("Error querying database", err)
		return -1, err
	}

	counter := 0
	for rows.Next() {
		schemalVer := new(dbMigratorSchema)
		err = rows.Scan(&schemalVer.id, &schemalVer.version, &schemalVer.timestamp)
		if err != nil {
			return -1, err
		}
		counter = counter + 1
		if counter == 1 {
			return schemalVer.version, nil
		}
	}

	return 0, nil
}

func readMigrationDirectory(dirPath string) (map[int]string, error) {

	prefix := "__"
	postfix := ".sql"
	versionFileMap := make(map[int]string)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		//files = append(files, path)
		if !info.IsDir() {
			fmt.Println(path)
			index1 := strings.Index(path, prefix) + len(prefix)
			index2 := strings.Index(path, postfix)
			if index1 == -1 || index2 == -1 {
				err = errors.New("Invalid file name found")
				return err
			}
			version, err := strconv.Atoi(path[index1:index2])
			if index1 == -1 || index2 == -1 {
				err = errors.New("Invalid file name found")
				return err
			}
			log.Printf("Found %s with version %d", path, version)
			versionFileMap[version-1] = path

		}
		return nil

	})
	if err != nil {
		panic(err)
	}

	return versionFileMap, err
}

func applyMigrations(db *sql.DB, versionFileMap *map[int]string, currentVersion int) error {

	//Sort the map
	keys := make([]int, 0, len(*versionFileMap))

	for k := range *versionFileMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	//iterate through map and apply the missing versions
	for _, k := range keys {

		fileVersion := k + 1
		filePath := (*versionFileMap)[k]
		if fileVersion > currentVersion {
			log.Printf("Applying migration for version %d from file %s", fileVersion, filePath)
			err := runMigration(db, filePath, fileVersion)
			if err != nil {
				log.Printf("Migration failed for %s to version %d, current version remains %d", filePath, fileVersion, currentVersion)
				return err
			}

		} else {
			log.Printf("Not applying already existing version %d skipping file %s", fileVersion, filePath)
		}
	}

	return nil

}

func runMigration(db *sql.DB, filePath string, version int) error {
	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		log.Printf("Error reading filr %s due to %s", filePath, err)
		return err
	}

	requests := strings.Split(string(file), ";")

	tx, err := db.Begin()

	if err != nil {
		log.Printf("Error starting transaction due to %s", err)
		return err
	}

	for _, request := range requests {
		if len(request) > 0 {
			_, err := tx.Exec(request)
			if err != nil {
				log.Printf("Transaction will rollback. Error during tx.Exec for file %s due to %s.", filePath, err)
				tx.Rollback()
				return err
			}
		}
	}

	nanos := time.Now().UnixNano()
	epochMillis := nanos / 1000000

	insertSQL := insertMigrationVersion()
	_, err = tx.Exec(insertSQL, version, epochMillis, filePath)
	if err != nil {
		log.Printf("Transaction will rollback. Error during tx.Exec for version update due to %v", err)
		tx.Rollback()
		return err
	}
	log.Printf("dbmigration_versioning update to %d %d", version, epochMillis)

	err = tx.Commit()
	if err != nil {
		log.Printf("Error during tx.Commit() for file %s due to %s.", filePath, err)
	} else {
		log.Printf("Successfully applied migration for %s to version %d", filePath, version)
	}

	return err

}
