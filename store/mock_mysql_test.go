package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/hewen/mastiff-go/config/storeconf"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitMockMysql(t *testing.T) {
	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)
}

func TestInitMockMysql_EmptyDir(t *testing.T) {
	db, err := InitMockMysql("")
	assert.NotNil(t, db)
	assert.Nil(t, err)
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
	initMysqlFunc = func(_ storeconf.MysqlConfig, _ ...DatabaseOption) (*DB, error) {
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

func TestInitMockMysql_GetFreePortError(t *testing.T) {
	orig := getFreePortFunc
	getFreePortFunc = func() (int, error) {
		return 0, fmt.Errorf("get port failed")
	}
	defer func() { getFreePortFunc = orig }()

	_, err := InitMockMysql("/tmp")
	assert.Error(t, err)
	assert.EqualError(t, err, "get port failed")
}

func TestInitMockMysql_StartServerError(t *testing.T) {
	orig := startMockMysqlFunc
	startMockMysqlFunc = func(_ string, _ *sqle.Engine, _ *memory.DbProvider) error {
		return fmt.Errorf("start server failed")
	}
	defer func() { startMockMysqlFunc = orig }()

	_, err := InitMockMysql("/tmp")
	assert.Error(t, err)
	assert.EqualError(t, err, "start server failed")
}

func TestInitMockMysql_InitMysqlError(t *testing.T) {
	orig := initMysqlFunc
	initMysqlFunc = func(_ storeconf.MysqlConfig, _ ...DatabaseOption) (*DB, error) {
		return nil, fmt.Errorf("init mysql failed")
	}
	defer func() { initMysqlFunc = orig }()

	_, err := InitMockMysql("/tmp")
	assert.Error(t, err)
	assert.EqualError(t, err, "init mysql failed")
}

func TestInitMockMysql_LoadSQLFileError(t *testing.T) {
	orig := loadSQLFilesFunc
	loadSQLFilesFunc = func(_ *DB, _ string) error {
		return fmt.Errorf("load sql failed")
	}
	defer func() { loadSQLFilesFunc = orig }()

	_, err := InitMockMysql("/tmp")
	assert.Error(t, err)
	assert.EqualError(t, err, "load sql failed")
}

func TestLoadSQLFiles_ReadDirError(t *testing.T) {
	err := loadSQLFiles(&DB{}, "/not/existing/path")
	assert.Error(t, err)
}

func TestLoadSQLFiles_OpenFileError(t *testing.T) {
	tmpDir := t.TempDir()

	badFile := filepath.Join(tmpDir, "test.sql")
	f, err := os.OpenFile(badFile, os.O_CREATE|os.O_WRONLY, 0222) // #nosec
	require.NoError(t, err)
	_ = f.Close()

	db, _, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	err = loadSQLFiles(&DB{DB: sqlxDB}, tmpDir)
	assert.Error(t, err)
}

func TestLoadSQLFiles_ExecError(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.sql"), []byte("INVALID SQL"), 0600)
	assert.NoError(t, err)

	db, _, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockDB := &DB{DB: sqlxDB}

	err = loadSQLFiles(mockDB, tmpDir)
	assert.Error(t, err)
}

func TestStartMockMysqlServer_NewServerError(t *testing.T) {
	orig := newMysqlServerFunc
	newMysqlServerFunc = func(
		_ server.Config,
		_ *sqle.Engine,
		_ sql.ContextFactory,
		_ server.SessionBuilder,
		_ server.ServerEventListener,
	) (*server.Server, error) {
		return nil, fmt.Errorf("new server fail")
	}

	defer func() { newMysqlServerFunc = orig }()

	engine, provider := createMockMysqlEngine("mockdb")
	err := startMockMysqlServer("localhost:3307", engine, provider)
	assert.Error(t, err)
	assert.EqualError(t, err, "new server fail")
}
