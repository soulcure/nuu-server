package models

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type PlatformAccount struct {
	Url         string `yaml:"url"`
	TransSerial string `yaml:"trans_serial"`
	Login       string `yaml:"login"`
	AuthCode    string `yaml:"auth_code"`

	// platform api config
	UsedTodayDetail string `yaml:"used_today_detail"`
	PackageSale     string `yaml:"package_sale"`
}

var (
	account PlatformAccount
)

func init() {
	path := "./conf/platform.yml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logrus.Panic("platform conf file does not exist")
	}

	data, _ := ioutil.ReadFile(path)
	if err := yaml.Unmarshal(data, &account); err != nil {
		logrus.Panic("platform conf yaml Unmarshal error ")
	}
}

func PackageDetailToday(ctx iris.Context) {
	deviceSn := ctx.FormValue("deviceSn")

	data := make(url.Values)
	data["itf_name"] = []string{account.UsedTodayDetail}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}

	rsp, err := http.PostForm(account.Url, data)
	if err == nil {
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			if _, err := ctx.Write(body); err != nil {
				logrus.Error(err)
			}
			return
		}
	}

	defer func() {
		if err = rsp.Body.Close(); err != nil {
			logrus.Error("http resp body close err:", err)
		}
	}()

	var res ProtocolRsp
	res.Code = ReqPlatformErrCode
	res.Msg = err.Error()
	res.ResponseWriter(ctx)

}

func QueryPackageForSale(ctx iris.Context) {
	deviceSn := ctx.FormValue("deviceSn")

	data := make(url.Values)
	data["itf_name"] = []string{account.PackageSale}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}

	rsp, err := http.PostForm(account.Url, data)
	if err == nil {
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			if _, err := ctx.Write(body); err != nil {
				logrus.Error(err)
			}
			return
		}
	}

	defer func() {
		if err = rsp.Body.Close(); err != nil {
			logrus.Error("http resp body close err:", err)
		}
	}()

	var res ProtocolRsp
	res.Code = ReqPlatformErrCode
	res.Msg = err.Error()
	res.ResponseWriter(ctx)

}
