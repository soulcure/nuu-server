package main

import (
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"nuu-server/src/file"
	"nuu-server/src/routes"
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

	runner := iris.Addr(":8899")
	//runner := iris.TLS(":8899", "./conf/server.crt", "./conf/server.key") //https
	if err := app.Run(runner, config); err != nil {
		logrus.Error(err)
	}

}
