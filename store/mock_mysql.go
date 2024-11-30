package store

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/hewen/mastiff-go/util"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
)

func InitMockMysql(sqlDir string) (*DB, error) {
	if sqlDir == "" {
		return nil, fmt.Errorf("sql dir empty")
	}

	dbName := "mockdb"
	db := memory.NewDatabase(dbName)
	db.BaseDatabase.EnablePrimaryKeyIndexes()
	pro := memory.NewDBProvider(db)
	engine := sqle.NewDefault(pro)

	port, err := util.GetFreePort()
	if err != nil {
		return nil, err
	}
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("localhost:%d", port),
	}

	s, err := server.NewServer(config, engine, memory.NewSessionBuilder(pro), nil)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.Start(); err != nil {
			panic(err)
		}
	}()

	DbConn, err := InitMysql(MysqlConf{
		DataSourceName: fmt.Sprintf("root:@tcp(%s)/%s?charset=utf8mb4&parseTime=true&interpolateParams=true", config.Address, dbName),
	})
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		return nil, err
	}
	files := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, info)
	}

	for _, file := range files {
		if path.Ext(file.Name()) != ".sql" {
			continue
		}

		fi, err := os.Open(sqlDir + "/" + file.Name())
		if err != nil {
			return nil, err
		}
		defer fi.Close()
		br := bufio.NewReader(fi)
		var sqlStr string
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			if string(a) != strings.TrimPrefix(string(a), "---") {
				continue
			}
			sqlStr += string(a)
		}
		_, err = DbConn.Exec(sqlStr)
		if err != nil {
			return nil, err
		}
	}

	return DbConn, nil
}
