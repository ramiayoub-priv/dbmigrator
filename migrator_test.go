package migrator

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestCreateMigrationTablesIfNotExist(t *testing.T) {

	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sql := createMigrationTableIfNotExists()

	fmt.Println(sql)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sql)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = createMigratorTableIfNotExists(db)

	require.Equal(t, nil, err)

}

func TestCheckVersion(t *testing.T) {

	db, mock, err := sqlmock.New()
	columns := []string{"id", "migrator_version", "migrator_timestamp"}

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sql := selectMigrationVersion()
	mock.ExpectPrepare(sql)
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 2, 123123123))

	version, err := checkVersion(db)

	require.Equal(t, nil, err)
	require.Equal(t, 2, version)

}

func TestReadMigrationDirectory(t *testing.T) {

	m, err := readMigrationDirectory("./migrationtest")

	require.Equal(t, nil, err)
	require.Equal(t, "migrationtest/Somesqlfile__1.sql", m[0])
	require.Equal(t, "migrationtest/Anothersqlfile_2__2.sql", m[1])
	require.Equal(t, "migrationtest/thirdfile__3.sql", m[2])
}

func TestApplyMigrations(t *testing.T) {
	m, err := readMigrationDirectory("./migrationtest")

	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sql1 := `CREATE TABLE dbmigrator_testone
	(
		id SERIAL NOT NULL PRIMARY KEY,
		column1 INT NOT NULL,
		column2 INT NOT NULL,
		column3 VARCHAR(255) NOT NULL
	);`

	sql2 := "INSERT into dbmigrator_testone (column1,column2,column3) VALUES (20,21,'testing version 2')"
	sql3 := "INSERT into dbmigrator_testone (column1,column2,column3) VALUES (30,31,'testing version 3')"

	versionSQL := insertMigrationVersion()

	fmt.Println(sql1)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sql2)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(versionSQL)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sql3)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(versionSQL)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = applyMigrations(db, &m, 1)

	require.Equal(t, nil, err)

}
