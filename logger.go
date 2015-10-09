package eweb

import (
	"fmt"
	"log"
	"os"
)

var globalLogger *logger

type logger struct {
	Logger      *log.Logger
	ShowConsole bool
	Level       int
}

const (
	LOG int = iota
	DEBUG
	INFO
	WARN
	ERROR
)

func init() {
	config := GetConfig("default")
	file, err := os.Create("log/app.log")
	if err != nil {
		log.Fatalln("fail to create app.log file!")
	}
	globalLogger = &logger{
		Logger: log.New(file, "", log.LstdFlags|log.Llongfile),
	}
	globalLogger.Logger.SetFlags(log.LstdFlags)
	globalLogger.Level = config.GetInt("log", "logLevel", LOG)
	globalLogger.ShowConsole = config.GetBool("log", "logShowConsole", true)
}

func logOut(level int, format string, v ...interface{}) {
	if level >= globalLogger.Level {
		info := fmt.Sprintf(format, v...)
		var prex string
		switch level {
		case DEBUG:
			prex = "DEBUG"
		case INFO:
			prex = "INFO"
		case WARN:
			prex = "WARN"
		case ERROR:
			prex = "ERROR"
		default:
			prex = "LOG"
		}
		prex = " [" + prex + "] "
		if globalLogger.ShowConsole {
			log.Println(prex, info)
		}
		globalLogger.Logger.Println(prex, info)
	}
}

func LogDebug(format string, v ...interface{}) {
	logOut(DEBUG, format, v...)
}

func LogWarn(format string, v ...interface{}) {
	logOut(WARN, format, v...)
}

func LogError(format string, v ...interface{}) {
	logOut(ERROR, format, v...)
}

func LogInfo(format string, v ...interface{}) {
	logOut(INFO, format, v...)
}

func Log(format string, v ...interface{}) {
	logOut(LOG, format, v...)
}
