package main

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"log"
	"routes"
)

func init() {
	f := routes.NewLogFile()
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(f)
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	/*logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})*/
}

func main() {
	f := routes.NewLogFile()
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
