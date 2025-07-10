package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gchaincl/sqlhooks"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestDriver struct {
}

func TestRegisterDriver(_ *testing.T) {
	registerHookDriver("test", &TestDriver{}, &SQLHooks{})
}

func (t *TestDriver) Open(_ string) (driver.Conn, error) {
	return nil, nil
}

func TestTransact(t *testing.T) {
	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)

	defer func() {
		_ = db.Close()
	}()

	err = db.Transact(func(_ *sqlx.Tx) error {
		return nil
	})
	assert.Nil(t, err)

	err = db.Transact(func(_ *sqlx.Tx) error {
		return fmt.Errorf("error")
	})
	assert.NotNil(t, err)

	err = db.Transact(func(_ *sqlx.Tx) error {
		panic("test")
	})
	assert.NotNil(t, err)
}

func TestSelectContext(t *testing.T) {
	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)

	var ids []uint64
	err = db.SelectContext(context.TODO(), &ids, "SELECT id FROM test WHERE id > ? limit 10", 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, len(ids))
}

func TestGetContext(t *testing.T) {
	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)

	var id int64
	err = db.GetContext(context.TODO(), &id, "SELECT id FROM test WHERE id = ?", 1)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.EqualValues(t, 0, id)
}

func TestInitDB_WithoutHookDriver(t *testing.T) {
	db, err := InitMockMysql("./test/")
	assert.Nil(t, err)

	driverName := "mysql"
	dataSource := db.DataSourceName
	driver := mysql.MySQLDriver{}

	_, err = InitDB(driverName, dataSource, driver, DatabaseOption{
		RegisterHookDriver: false,
		Hook:               &SQLHooks{},
		MaxIdleConns:       1,
		MaxOpenConns:       1,
		ConnMaxLifetime:    time.Minute,
	})

	assert.Nil(t, err)
}

func TestInitDB_OpenError(t *testing.T) {
	_, err := InitDB("invalidDriver", "bad-dsn", nil)
	assert.Error(t, err)
}

func TestInitDB_PingError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.Nil(t, err)
	mock.ExpectPing().WillReturnError(fmt.Errorf("ping failed"))

	defer func() {
		_ = db.Close()
	}()

	sql.Register("mock", sqlhooks.Wrap(db.Driver(), &SQLHooks{}))
	_, err = InitDB("mock", "any-dsn", db.Driver(), DatabaseOption{
		RegisterHookDriver: false,
	})
	assert.Error(t, err)
}

func TestTransact_BeginError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = sqlDB.Close()
	}()

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
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = sqlDB.Close()
	}()

	db := sqlx.NewDb(sqlDB, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback()

	storeDB := &DB{DB: db}
	err = storeDB.Transact(func(_ *sqlx.Tx) error {
		return fmt.Errorf("expected error")
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "expected error")
}

func TestTransact_Panic(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = sqlDB.Close()
	}()

	db := sqlx.NewDb(sqlDB, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback()

	storeDB := &DB{DB: db}
	err = storeDB.Transact(func(_ *sqlx.Tx) error {
		panic("panic error")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic error")
}

func TestTransact_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = sqlDB.Close()
	}()

	db := sqlx.NewDb(sqlDB, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()

	storeDB := &DB{DB: db}
	err = storeDB.Transact(func(_ *sqlx.Tx) error {
		// do nothing
		return nil
	})
	assert.NoError(t, err)
}
