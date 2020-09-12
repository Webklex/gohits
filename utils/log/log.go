package log

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"time"
)

type Config struct {
	Silent bool
}

var Log = Config{Silent: false}

func (c *Config) log(_type string, a ...interface{}) {
	if c.Silent {
		return
	}

	_datetime := time.Now().Format("02.01.2006 15:04:05")
	message := fmt.Sprintf("[%s] [%s] %s", _type, _datetime, a[0])
	switch _type {
	case "error":
		color.Magenta(message)
	case "update":
		color.Cyan(message)
	case "success":
		color.Green(message)
	default:
		color.Cyan(message)
	}
}

func Success(v ...interface{}) {
	Log.log("success", v)
}

func Info(v ...interface{}) {
	Log.log("info", v)
}

func Error(v ...interface{}) {
	Log.log("error", v)
}

func Fatal(v ...interface{}) {
	Error(v)
	os.Exit(1)
}
