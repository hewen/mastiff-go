package store

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/gchaincl/sqlhooks"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

type DatabaseOption struct {
	RegisterHookDriver bool
}

func InitDB(driverName, dataSourceName string, driver driver.Driver, opt ...DatabaseOption) (*DB, error) {
	if len(opt) == 0 || opt[0].RegisterHookDriver {
		driverName += "WithHooks"
		registerHookDriver(driverName, driver)
	}

	conn, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(8)
	conn.SetMaxOpenConns(128)
	conn.SetConnMaxLifetime(time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{sqlx.NewDb(conn, driverName)}, nil
}

func registerHookDriver(driverName string, driver driver.Driver) {
	drivers := sql.Drivers()
	var registerHook bool
	for i := range drivers {
		if drivers[i] == driverName {
			registerHook = true
			break
		}
	}
	if !registerHook {
		sql.Register(driverName, sqlhooks.Wrap(driver, &Hooks{}))
	}
}

func (db *DB) Transact(fn func(*sqlx.Tx) error) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		p := recover()
		switch {
		case p != nil:
			_ = tx.Rollback()
			err = fmt.Errorf("%#v", p)
		case err != nil:
			_ = tx.Rollback()
		default:
			err = tx.Commit()
		}
	}()

	return fn(tx)
}

func InitMysql(conf MysqlConf, opt ...DatabaseOption) (*DB, error) {
	return InitDB("mysql", conf.DataSourceName, &mysql.MySQLDriver{}, opt...)
}
