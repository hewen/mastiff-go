// Package store mock mysql
package store

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	gsql "github.com/dolthub/go-mysql-server/sql"
	vsql "github.com/dolthub/vitess/go/mysql"
	"github.com/hewen/mastiff-go/config/storeconf"
	"github.com/hewen/mastiff-go/pkg/util"
)

var (
	getFreePortFunc    = util.GetFreePort
	startMockMysqlFunc = startMockMysqlServer
	initMysqlFunc      = InitMysql
	loadSQLFilesFunc   = loadSQLFiles
	newMysqlServerFunc = server.NewServer
)

// InitMockMysql initializes a MySQL connection with the given configuration.
func InitMockMysql(sqlDir string) (*DB, error) {
	// 1. create mock mysql engine
	dbName := "mockdb"
	engine, provider := createMockMysqlEngine(dbName)

	port, err := getFreePortFunc()
	if err != nil {
		return nil, err
	}

	// 2. start mock mysql server
	address := fmt.Sprintf("localhost:%d", port)
	err = startMockMysqlFunc(address, engine, provider)
	if err != nil {
		return nil, err
	}

	// 3. connect to mock mysql server
	connStr := fmt.Sprintf("root:@tcp(%s)/%s?charset=utf8mb4&parseTime=true&interpolateParams=true", address, dbName)
	dbConn, err := initMysqlFunc(storeconf.MysqlConfig{DataSourceName: connStr}, DatabaseOption{RegisterHookDriver: true})
	if err != nil {
		return nil, err
	}

	// 4. load sql files
	err = loadSQLFilesFunc(dbConn, sqlDir)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}
func createMockMysqlEngine(dbName string) (*sqle.Engine, *memory.DbProvider) {
	db := memory.NewDatabase(dbName)
	db.EnablePrimaryKeyIndexes()
	provider := memory.NewDBProvider(db)
	engine := sqle.NewDefault(provider)
	return engine, provider
}

func startMockMysqlServer(address string, engine *sqle.Engine, provider *memory.DbProvider) error {
	cfg := server.Config{
		Protocol: "tcp",
		Address:  address,
	}

	sessionBuilder := func(_ context.Context, c *vsql.Conn, addr string) (gsql.Session, error) {
		var host, user string
		mysqlConnectionUser, ok := c.UserData.(gsql.MysqlConnectionUser)
		if ok {
			host = mysqlConnectionUser.Host
			user = mysqlConnectionUser.User
		}
		client := gsql.Client{Address: host, User: user, Capabilities: c.Capabilities}
		return memory.NewSession(gsql.NewBaseSessionWithClientServer(addr, client, c.ConnectionID), provider), nil
	}

	srv, err := newMysqlServerFunc(cfg, engine, gsql.NewContext, sessionBuilder, nil)
	if err != nil {
		return err
	}

	go func() {
		if err := srv.Start(); err != nil {
			panic(err)
		}
	}()

	return nil
}

func loadSQLFiles(dbConn *DB, sqlDir string) error {
	if sqlDir == "" {
		return nil
	}

	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		file, err := os.Open(filepath.Join(sqlDir, entry.Name())) // #nosec G304 -- entry name from os.ReadDir, safe
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		scanner := bufio.NewScanner(file)
		var sb strings.Builder
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "---") {
				continue
			}
			_, err = sb.WriteString(scanner.Text())
			if err != nil {
				return err
			}
		}

		_, err = dbConn.Exec(sb.String())
		if err != nil {
			return err
		}
	}
	return nil
}
