// Package store mock mysql
package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/hewen/mastiff-go/util"
)

// InitMockMysql initializes a MySQL connection with the given configuration.
func InitMockMysql(sqlDir string) (*DB, error) {
	if sqlDir == "" {
		return nil, fmt.Errorf("sql dir empty")
	}

	// 1. 创建 engine + server 配置
	dbName := "mockdb"
	engine, provider := createMockMysqlEngine(dbName)

	port, err := util.GetFreePort()
	if err != nil {
		return nil, err
	}

	// 2. 启动服务
	address := fmt.Sprintf("localhost:%d", port)
	err = startMockMysqlServer(address, engine, provider)
	if err != nil {
		return nil, err
	}

	// 3. 初始化连接
	connStr := fmt.Sprintf("root:@tcp(%s)/%s?charset=utf8mb4&parseTime=true&interpolateParams=true", address, dbName)
	dbConn, err := InitMysql(MysqlConf{DataSourceName: connStr})
	if err != nil {
		return nil, err
	}

	// 4. 加载 SQL 文件
	err = loadSQLFiles(dbConn, sqlDir)
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
	config := server.Config{
		Protocol: "tcp",
		Address:  address,
	}

	srv, err := server.NewServer(config, engine, memory.NewSessionBuilder(provider), nil)
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
