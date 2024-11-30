package store

import (
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
