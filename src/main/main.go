package main

import (
	"file"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"routes"
)

func init() {
	f := file.NewLogFile()
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(f)

	//logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	f := file.NewLogFile()
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Printf("close log file error: %s", err)
		}
	}()

	// the rest of the code stays the same.
	app := iris.New()
	routes.Hub(app)

	app.Logger().SetOutput(f)
	app.Logger().SetLevel("debug")

	config := iris.WithConfiguration(iris.YAML("./conf/config.yml"))

	if err := app.Run(iris.Addr(":8899"), config); err != nil {
		logrus.Error(err)
	}

}
