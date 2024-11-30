package store

type (
	MysqlConf struct {
		DataSourceName     string
		RegisterHookDriver bool
	}

	RedisConf struct {
		Addr               string
		Password           string
		DB                 int
		RegisterHookDriver bool
	}
)
