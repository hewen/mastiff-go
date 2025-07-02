package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type TestDriver struct {
}

func TestRegisterDriver(_ *testing.T) {
	registerHookDriver("test", &TestDriver{})
}

func (t *TestDriver) Open(_ string) (driver.Conn, error) {
	return nil, nil
}

func TestTransact(t *testing.T) {
	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)

	err = db.Transact(func(_ *sqlx.Tx) error {
		return nil
	})
	assert.Nil(t, err)

	err = db.Transact(func(_ *sqlx.Tx) error {
		return fmt.Errorf("error")
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
