package main

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"routes"
	"time"
)

func init() {
	f := newLogFile()
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(f)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

// Get a filename based on the date, just for the sugar.
func todayFilename() string {
	today := time.Now().Format("2006-01-02")
	return today + ".txt"
}

func newLogFile() *os.File {
	filename := todayFilename()
	// Open the file, this will append to the today's file if server restarted.
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	return f
}

func main() {
	f := newLogFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("close log file error: %s", err)
		}
	}()

	// the rest of the code stays the same.
	app := iris.New()
	routes.Hub(app)

	app.Logger().SetOutput(f)
	app.Logger().SetLevel("debug")

	config := iris.WithConfiguration(iris.YAML("./conf/iris.yml"))

	if err := app.Run(iris.Addr(":8899"), config); err != nil {
		logrus.Error(err)
	}

}
