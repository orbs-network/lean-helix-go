package logger

type TestLog interface {
	Fatal(args ...interface{})
	Log(args ...interface{})
	Error(args ...interface{})
	Name() string
}
