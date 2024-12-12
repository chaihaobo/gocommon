package main

import (
	"context"

	"github.com/chaihaobo/gocommon/logger"
)

func main() {
	l, _, _ := logger.New(logger.Config{
		Encoding: "jsoncolor",
	})
	l.Info(context.Background(), "hello world")

}
