package logger

type Config struct {
	Level   LogLevel
	Output  string
	MaxSize int
}
