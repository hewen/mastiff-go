package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
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
	driverName := "mysql"
	dataSource := "root:password@tcp(127.0.0.1:3306)/test"
	driver := mysql.MySQLDriver{}

	// 使用 RegisterHookDriver = false
	db, err := InitDB(driverName, dataSource, driver, DatabaseOption{
		RegisterHookDriver: false,
		Hook:               &SQLHooks{},
		MaxIdleConns:       1,
		MaxOpenConns:       1,
		ConnMaxLifetime:    time.Minute,
	})

	// 因为你未实际连通 mysql，此处我们只验证分支和错误返回
	assert.Nil(t, db)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "dial") // 或其他 mysql 连接错误
}
