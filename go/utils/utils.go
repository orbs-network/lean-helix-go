package utils

import (
	"fmt"
)

// TODO think of something better here - this Logger wrapper is just so we don't have to change all the logger calls when changing logger impl

type ConsoleLogger struct{}

var Logger = new(ConsoleLogger)

type BasicLogger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func (log *ConsoleLogger) Info(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	fmt.Println(str)
}

func (log *ConsoleLogger) Error(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	fmt.Println(str)
}
