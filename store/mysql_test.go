package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/hewen/mastiff-go/config/storeconf"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockStoreDB(t *testing.T) (*DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	db := sqlx.NewDb(sqlDB, "sqlmock")
	return &DB{
		DB:             db,
		DataSourceName: "mock-dsn",
	}, mock
}

type TestDriver struct{}

func (t *TestDriver) Open(_ string) (driver.Conn, error) {
	return nil, nil
}

type ErrorDriver struct{}

func (e *ErrorDriver) Open(_ string) (driver.Conn, error) {
	return nil, fmt.Errorf("open connection failed")
}

func TestRegisterHookDriver(t *testing.T) {
	assert.NotPanics(t, func() {
		registerHookDriver("test-driver", &TestDriver{}, &SQLHooks{})
	})
}

func TestInitDB_OpenError(t *testing.T) {
	_, err := InitDB("invalid-driver", "bad-dsn", nil)
	assert.Error(t, err)
}

func TestTransact_BeginError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin failed"))

	db := sqlx.NewDb(sqlDB, "sqlmock")
	storeDB := &DB{DB: db}

	err = storeDB.Transact(func(_ *sqlx.Tx) error {
		return nil
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "begin failed")
}

func TestTransact_FuncReturnsError(t *testing.T) {
	db, mock := newMockStoreDB(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := db.Transact(func(_ *sqlx.Tx) error {
		return fmt.Errorf("expected error")
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "expected error")
}

func TestTransact_Panic(t *testing.T) {
	db, mock := newMockStoreDB(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := db.Transact(func(_ *sqlx.Tx) error {
		panic("panic error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic error")
}

func TestTransact_Success(t *testing.T) {
	db, mock := newMockStoreDB(t)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := db.Transact(func(_ *sqlx.Tx) error {
		return nil
	})

	assert.NoError(t, err)
}

func TestSelectContext(t *testing.T) {
	db, mock := newMockStoreDB(t)

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT id FROM test").
		WithArgs(1).
		WillReturnRows(rows)

	var ids []uint64
	err := db.SelectContext(
		context.TODO(),
		&ids,
		"SELECT id FROM test WHERE id > ? limit 10",
		1,
	)

	assert.NoError(t, err)
	assert.Len(t, ids, 0)
}

func TestGetContext_NoRows(t *testing.T) {
	db, mock := newMockStoreDB(t)

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT id FROM test").
		WithArgs(1).
		WillReturnRows(rows)

	var id int64
	err := db.GetContext(
		context.TODO(),
		&id,
		"SELECT id FROM test WHERE id = ?",
		1,
	)

	assert.Equal(t, sql.ErrNoRows, err)
	assert.EqualValues(t, 0, id)
}

func TestInitMysql_DriverSmoke(t *testing.T) {
	_, _ = mysql.MySQLDriver{}, sql.ErrNoRows
}

func TestClose(t *testing.T) {
	db, mock := newMockStoreDB(t)
	mock.ExpectClose()
	err := db.Close()
	assert.NoError(t, err)
}

func TestInitMysql(t *testing.T) {
	_, err := InitMysql(storeconf.MysqlConfig{DataSourceName: "invalid-dsn:test"})
	assert.Error(t, err)
}

func TestSQLHooks_Before(t *testing.T) {
	hooks := &SQLHooks{}
	ctx, err := hooks.Before(context.Background(), "SELECT 1", 1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
}

func TestSQLHooks_After(t *testing.T) {
	hooks := &SQLHooks{}
	ctx := context.Background()
	ctx, err := hooks.After(ctx, "SELECT 1", 1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
}

func TestInitDB_DefaultOptions(t *testing.T) {
	_, err := InitDB("invalid-driver-default", "dsn", nil)
	assert.Error(t, err)
}

func TestInitDB_WithHookDriver(t *testing.T) {
	driverName := "error-driver-hook-test"
	_, err := InitDB(driverName, "dsn", &ErrorDriver{}, DatabaseOption{RegisterHookDriver: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open connection failed")
}

func TestInitDB_WithOptions(t *testing.T) {
	_, err := InitDB("invalid-driver-opts", "dsn", nil, DatabaseOption{
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 2 * time.Minute,
	})
	assert.Error(t, err)
}

func TestIsDriverRegistered_Registered(t *testing.T) {
	assert.True(t, isDriverRegistered("sqlmock"))
}

func TestIsDriverRegistered_NotRegistered(t *testing.T) {
	assert.False(t, isDriverRegistered("non-existent-driver-12345"))
}
