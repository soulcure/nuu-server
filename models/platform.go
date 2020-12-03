package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"nuu-server/cpython"
	"nuu-server/file"
	"nuu-server/mysql"
	"nuu-server/paypal"
	"nuu-server/redis"
	"nuu-server/utils"
	"os"

	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//PlatformAccount 平台账号
type PlatformAccount struct {
	URL         string `yaml:"url"`
	TransSerial string `yaml:"trans_serial"`
	Login       string `yaml:"login"`
	AuthCode    string `yaml:"auth_code"`

	//PayPal account
	ClientID    string `yaml:"clientID"`
	SecretID    string `yaml:"secretID"`
	APIBaseLive string `yaml:"APIBaseLive"`

	// platform api config
	UsedTodayDetail string `yaml:"used_today_detail"` //今天使用情况
	PackageSale     string `yaml:"package_sale"`      //购买的套餐列表
	PackagePayDone  string `yaml:"package_pay_done"`  //购买套餐
	PackageQuery    string `yaml:"package_query"`     //所有流量包情况
	SetupWifi       string `yaml:"setup_wifi"`        //设置wifi 名称和密码
	UsedDetail      string `yaml:"used_detail"`       //查询使用详情

	SMTP         string `yaml:"smtp"`          //邮箱地址 类型
	SendAccount  string `yaml:"send_account"`  //发件人账号
	SendPassword string `yaml:"send_password"` //发件人密码
}

type BuyPackageResult struct {
	ItfName             string `json:"itf_name"`
	TransSerial         string `json:"trans_serial"`
	ErrCode             int    `json:"err_code"`
	ErrDesc             string `json:"err_desc"`
	DevicePackageId     int    `json:"device_package_id,omitempty"`
	DevicePackageIdList string `json:"device_package_id_list,omitempty"`
	OrderId             string `json:"order_id,omitempty"`
}

var (
	account PlatformAccount
)

func init() {
	path := "./conf/config.yml"
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

	rsp, err := http.PostForm(account.URL, data)
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

	var res ErrorRsp
	res.Code = ReqPlatformErrCode
	res.Message = err.Error()
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

func QueryPackageForSale(ctx iris.Context) {
	deviceSn := ctx.FormValue("deviceSn")
	/*if body, err := redis.GetBytes(deviceSn); err == nil {
		if _, err := ctx.Write(body); err != nil {
			logrus.Error("QueryPackageForSale error", err)
		}
		return
	}*/

	data := make(url.Values)
	data["itf_name"] = []string{account.PackageSale}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}

	rsp, err := http.PostForm(account.URL, data)
	if err == nil {
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			//if _, err := redis.SetBytes(deviceSn, body); err == nil {
			if _, err := ctx.Write(body); err != nil {
				logrus.Error(err)
			}
			return
			//}
		}
	}

	defer func() {
		if err = rsp.Body.Close(); err != nil {
			logrus.Error("http resp body close err:", err)
		}
	}()

	var res ErrorRsp
	res.Code = ReqPlatformErrCode
	res.Message = err.Error()
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

func PayPalDone(ctx iris.Context, order *mysql.OrderReq) {
	data := make(url.Values)
	data["itf_name"] = []string{account.PackagePayDone}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{order.DeviceSn}
	data["package_id"] = []string{fmt.Sprintf("%d", order.PackageId)}
	data["begin_date"] = []string{order.BeginDate}
	data["total_num"] = []string{fmt.Sprintf("%d", order.Count)}
	data["valid_type"] = []string{fmt.Sprintf("%d", order.EffectiveType)}

	var (
		resp      *http.Response
		err       error
		body      []byte
		buyResult BuyPackageResult
	)

	defer func() {
		if err = resp.Body.Close(); err != nil {
			logrus.Error("http resp body close err:", err)
		}
	}()

	if resp, err = http.PostForm(account.URL, data); err == nil {
		if body, err = ioutil.ReadAll(resp.Body); err == nil {
			logrus.Debug("platform buy_package resp body :", string(body))

			if err = json.Unmarshal(body, &buyResult); err == nil {
				if buyResult.ErrCode == 0 {
					order.Effective = 1 //流量包已经生效

					pOrder := &mysql.BuyPackagePlatform{
						UserId:   order.UserId,
						Uuid:     order.Uuid,
						DeviceSn: order.DeviceSn,

						PackageId:   order.PackageId,
						PackageName: order.PackageName,
						Currency:    order.Currency,
						Count:       order.Count,
						Money:       order.Money,
						OrderTime:   order.OrderTime,

						PlatformOrderId:     buyResult.OrderId,
						DevicePackageId:     buyResult.DevicePackageId,
						DevicePackageIdList: buyResult.DevicePackageIdList,
					}

					if err = mysql.UpdateOrderTX(order, pOrder); err == nil {
						if _, err = redis.DelKey(order.OrderId); err == nil {
							var res ProtocolRsp
							res.Data = buyResult
							res.ResponseWriter(ctx, http.StatusOK)
							return
						}
					}
				} else {
					err = errors.New(buyResult.ErrDesc)
				}
			}
		}
	}

	var res ErrorRsp
	res.Code = ReqPlatformErrCode
	if err != nil {
		res.Message = err.Error()
	} else {
		res.Message = PackagePlatformErr
	}
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

func OrderPay(ctx iris.Context) {
	paymentId := ctx.FormValue("paymentId")
	orderId := ctx.FormValue("orderId")
	// Create a client instance
	if c, err := paypal.NewClient(account.ClientID, account.SecretID, account.APIBaseLive); err == nil {
		c.SetLog(file.NewLogFile()) // Set log to terminal stdout

		if _, err := c.GetAccessToken(); err == nil {
			if payment, err := c.GetPayment(paymentId); err == nil {
				logrus.Debug("payment:", payment)

				orderBean := &mysql.OrderReq{}
				if err := redis.GetStruct(orderId, orderBean); err == nil {
					orderBean.Status = 1        //订单已经支付
					orderBean.PayId = paymentId //支付订单号
					PayPalDone(ctx, orderBean)
				}

			} else {
				logrus.Error("payment err:", err)
			}
		} else {
			logrus.Error("GetAccessToken err:", err)
		}
	}

}

func PackageQuery(ctx iris.Context) {
	deviceSn := ctx.FormValue("deviceSn")

	data := make(url.Values)
	data["itf_name"] = []string{account.PackageQuery}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}

	rsp, err := http.PostForm(account.URL, data)
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

	var res ErrorRsp
	res.Code = ReqPlatformErrCode
	res.Message = err.Error()
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

func SettWifiPassword(ctx iris.Context) {
	deviceSn := ctx.FormValue("deviceSn")
	name := ctx.FormValue("name")
	password := ctx.FormValue("password")

	data := make(url.Values)
	data["itf_name"] = []string{account.SetupWifi}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}
	data["ssid"] = []string{name}
	data["wifi_password"] = []string{password}

	rsp, err := http.PostForm(account.URL, data)
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

	var res ErrorRsp
	res.Code = ReqPlatformErrCode
	res.Message = err.Error()
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

func UsedDetailPeriod(ctx iris.Context) {
	deviceSn := ctx.FormValue("deviceSn")
	beginDate := ctx.FormValue("beginDate")
	endDate := ctx.FormValue("endDate")

	data := make(url.Values)
	data["itf_name"] = []string{account.UsedDetail}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}
	data["begin_date"] = []string{beginDate}
	data["end_date"] = []string{endDate}

	rsp, err := http.PostForm(account.URL, data)
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

	var res ErrorRsp
	res.Code = ReqPlatformErrCode
	res.Message = err.Error()
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

//找回密码
func SendPasswordEmail(ctx iris.Context) {
	email := ctx.FormValue("email")

	if email == "" {
		var res ErrorRsp
		res.Code = AccountErrCode
		res.Message = EmailEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return
	}

	if !utils.IsEmail(email) {
		var res ErrorRsp
		res.Code = AccountErrCode
		res.Message = EmailFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return
	}

	if password, err := mysql.AccountByEmail(email); err == nil {
		if err := cpython.SendEmail(account.SMTP, account.SendAccount, account.SendPassword, email, password); err == nil {
			var res ProtocolRsp
			res.Data = password
			res.ResponseWriter(ctx, http.StatusBadRequest)
		} else {
			var res ErrorRsp
			res.Code = AccountErrCode
			res.Message = err.Error()
			res.ResponseWriter(ctx, http.StatusBadRequest)
		}

	} else {
		var res ErrorRsp
		res.Code = AccountErrCode
		res.Message = err.Error()
		res.ResponseWriter(ctx, http.StatusBadRequest)
	}
}
