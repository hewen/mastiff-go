package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitMockMysql(t *testing.T) {
	_, err := InitMockMysql("")
	assert.NotNil(t, err)

	db, err := InitMockMysql("./test/")
	assert.NotNil(t, db)
	assert.Nil(t, err)
}
