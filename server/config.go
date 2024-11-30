package server

type (
	HTTPConfig struct {
		Addr         string
		TimeoutRead  int64
		TimeoutWrite int64
	}

	GrpcConfig struct {
		Addr    string
		Timeout int64
	}
)
