package eweb

import (
	"fmt"
	"log"
	"os"
)

var _logger *logger

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
	file, err := os.Create("log/app.log")
	if err != nil {
		log.Fatalln("fail to create app.log file!")
	}
	_logger = &logger{
		Logger: log.New(file, "", log.LstdFlags|log.Llongfile),
	}
	_logger.Logger.SetFlags(log.LstdFlags)
	config := MustGetConfig("default")
	if config != nil {
		_logger.Level = config.GetInt("logLevel")
		_logger.ShowConsole = config.GetBool("logShowConsole")
	}
}

func logOut(level int, format string, v ...interface{}) {
	if level >= _logger.Level {
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
		if _logger.ShowConsole {
			log.Println(prex, info)
		}
		_logger.Logger.Println(prex, info)
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
