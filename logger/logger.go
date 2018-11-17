package logger

import (
	"os"
	filetool "psDicountsBot/file"
	"strconv"
	"time"
)

var logFile *os.File
var logInited bool = false
var LOG_FILENAME = "bot_log.log"

func Init() {
	logFile = filetool.CreateFile(LOG_FILENAME)
	filetool.AppendToFile(logFile, getDateTime()+"Log started\n")
	filetool.CloseFile(logFile)
	logInited = true
}

func Write(logMessage string) {
	if logInited {
		logFile = filetool.OpenFile(LOG_FILENAME)
		filetool.AppendToFile(logFile, getDateTime()+logMessage+"\n")
		filetool.CloseFile(logFile)
	} else {
		Init()
		logFile = filetool.OpenFile(LOG_FILENAME)
		filetool.AppendToFile(logFile, getDateTime()+logMessage+"\n")
		filetool.CloseFile(logFile)
	}
}

func getDateTime() string {
	var hour = time.Now().Hour()
	var minute = time.Now().Minute()
	var second = time.Now().Second()
	var yy, mm, dd = time.Now().Date()
	return strconv.Itoa(dd) + "." + mm.String() + "." + strconv.Itoa(yy) + "_" + strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + ":" + strconv.Itoa(second) + "_|_"
}
