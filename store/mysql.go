// Package store mysql database
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

// DB represents a database connection with sqlx capabilities.
type DB struct {
	*sqlx.DB
	DataSourceName string
}

// DatabaseOption defines options for initializing a database connection.
type DatabaseOption struct {
	RegisterHookDriver bool
	Hook               sqlhooks.Hooks
	MaxIdleConns       int
	MaxOpenConns       int
	ConnMaxLifetime    time.Duration
}

// InitDB initializes a database connection.
func InitDB(driverName, dataSourceName string, driver driver.Driver, opt ...DatabaseOption) (*DB, error) {
	var optCfg DatabaseOption
	if len(opt) > 0 {
		optCfg = opt[0]
	}
	if optCfg.RegisterHookDriver {
		driverName += "WithHooks"
		if optCfg.Hook == nil {
			optCfg.Hook = &SQLHooks{}
		}
		registerHookDriver(driverName, driver, optCfg.Hook)
	}

	if optCfg.MaxIdleConns == 0 {
		optCfg.MaxIdleConns = 8
	}
	if optCfg.MaxOpenConns == 0 {
		optCfg.MaxOpenConns = 128
	}
	if optCfg.ConnMaxLifetime == 0 {
		optCfg.ConnMaxLifetime = time.Minute
	}

	conn, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(optCfg.MaxIdleConns)
	conn.SetMaxOpenConns(optCfg.MaxOpenConns)
	conn.SetConnMaxLifetime(optCfg.ConnMaxLifetime)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{DB: sqlx.NewDb(conn, driverName), DataSourceName: dataSourceName}, nil
}

// registerHookDriver registers a driver with sqlhooks if it is not already registered.
func registerHookDriver(driverName string, driver driver.Driver, hook sqlhooks.Hooks) {
	if !isDriverRegistered(driverName) {
		sql.Register(driverName, sqlhooks.Wrap(driver, hook))
	}
}

// isDriverRegistered check driver registered.
func isDriverRegistered(name string) bool {
	for _, driverName := range sql.Drivers() {
		if driverName == name {
			return true
		}
	}
	return false
}

// Transact executes a function within a transaction, handling commit and rollback automatically.
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

// Close db conn.
func (db *DB) Close() error {
	return db.DB.Close()
}

// InitMysql initializes a MySQL connection.
func InitMysql(conf MysqlConf, opt ...DatabaseOption) (*DB, error) {
	return InitDB("mysql", conf.DataSourceName, mysql.MySQLDriver{}, opt...)
}
