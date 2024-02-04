package utils

import (
	"os"

	"github.com/withmandala/go-log"
)

var Logger *log.Logger

func InitLogger() {
	isProduction, ok := os.LookupEnv("IS_PRODUCTION")
	if !ok || (isProduction != "true" && isProduction != "false") {
		isProduction = "false"
	}
	if isProduction == "true" {
		Logger = log.New(os.Stdout).WithColor().WithoutDebug().WithTimestamp()
		return
	}

	Logger = log.New(os.Stdout).WithColor().WithDebug().WithTimestamp()
}
