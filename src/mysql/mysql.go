package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type DBConfig struct {
	Mysql DBInfo `yaml:"mysql"`
}

//数据库账号配置
type DBInfo struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

var (
	db       *sql.DB
	dbConfig DBConfig
)

func init() {
	path := "./conf/config.yml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("db conf file does not exist")
	}

	data, _ := ioutil.ReadFile(path)
	if err := yaml.Unmarshal(data, &dbConfig); err != nil {
		logrus.Panic("db conf yaml Unmarshal error ")
	}

	dbName := getConnURL(&dbConfig.Mysql)

	database, err := sql.Open("mysql", dbName)
	if err != nil {
		logrus.Panic("mysql can not connect")
		return
	}
	db = database
	logrus.Print("mysql connect at ", dbName)
}

func getConnURL(info *DBInfo) (url string) {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		info.User, info.Password, info.Host, info.Port, info.Database, info.Charset)
}
