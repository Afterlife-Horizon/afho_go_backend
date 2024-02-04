package utils

import (
	"os"

	"github.com/withmandala/go-log"
)

var Logger *log.Logger

func InitLogger() {
	if os.Getenv("ENV") == "production" {
		Logger = log.New(os.Stdout).WithColor().WithoutDebug().WithTimestamp()
		return
	}

	Logger = log.New(os.Stdout).WithColor().WithDebug().WithTimestamp()
}
