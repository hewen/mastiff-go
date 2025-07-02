package store

import (
	"errors"
	"testing"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/stretchr/testify/assert"
)

func TestInitMockMysql(t *testing.T) {
	_, err := InitMockMysql("")
	assert.NotNil(t, err)

	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)
}

func TestInitMockMysql_EmptyDir(t *testing.T) {
	db, err := InitMockMysql("")
	assert.Nil(t, db)
	assert.EqualError(t, err, "sql dir empty")
}

func TestInitMockMysql_ErrorDir(t *testing.T) {
	db, err := InitMockMysql("error")
	assert.Nil(t, db)
	assert.NotNil(t, err)
}

func TestInitMockMysql_FreePortFail(t *testing.T) {
	original := getFreePortFunc
	getFreePortFunc = func() (int, error) {
		return 0, errors.New("mock free port error")
	}
	defer func() { getFreePortFunc = original }()

	db, err := InitMockMysql("testdata/sql")
	assert.Nil(t, db)
	assert.EqualError(t, err, "mock free port error")
}

func TestInitMockMysql_StartServerFail(t *testing.T) {
	original := startMockMysqlFunc
	startMockMysqlFunc = func(_ string, _ *sqle.Engine, _ *memory.DbProvider) error {
		return errors.New("mock start server fail")
	}
	defer func() { startMockMysqlFunc = original }()

	db, err := InitMockMysql("testdata/sql")
	assert.Nil(t, db)
	assert.EqualError(t, err, "mock start server fail")
}

func TestInitMockMysql_InitMysqlFail(t *testing.T) {
	original := initMysqlFunc
	initMysqlFunc = func(_ MysqlConf, _ ...DatabaseOption) (*DB, error) {
		return nil, errors.New("mock init mysql error")
	}
	defer func() { initMysqlFunc = original }()

	db, err := InitMockMysql("testdata/sql")
	assert.Nil(t, db)
	assert.EqualError(t, err, "mock init mysql error")
}

func TestInitMockMysql_LoadSQLFail(t *testing.T) {
	original := loadSQLFilesFunc
	loadSQLFilesFunc = func(_ *DB, _ string) error {
		return errors.New("mock load sql error")
	}
	defer func() { loadSQLFilesFunc = original }()

	db, err := InitMockMysql("testdata/sql")
	assert.Nil(t, db)
	assert.EqualError(t, err, "mock load sql error")
}
