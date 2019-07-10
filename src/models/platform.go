package models

import (
	"encoding/json"
	"errors"
	"file"
	"fmt"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"mysql"
	"net/http"
	"net/url"
	"os"
	"paypal"
	"redis"
)

type PlatformAccount struct {
	Url         string `yaml:"url"`
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
	if body, err := redis.GetBytes(deviceSn); err == nil {
		if _, err := ctx.Write(body); err != nil {
			logrus.Error("QueryPackageForSale error", err)
		}
		return
	}

	data := make(url.Values)
	data["itf_name"] = []string{account.PackageSale}
	data["trans_serial"] = []string{account.TransSerial}
	data["login"] = []string{account.Login}
	data["auth_code"] = []string{account.AuthCode}
	data["device_sn"] = []string{deviceSn}

	rsp, err := http.PostForm(account.Url, data)
	if err == nil {
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			if _, err := redis.SetBytes(deviceSn, body); err == nil {
				if _, err := ctx.Write(body); err != nil {
					logrus.Error(err)
				}
				return
			}
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

	if resp, err = http.PostForm(account.Url, data); err == nil {
		if body, err = ioutil.ReadAll(resp.Body); err == nil {
			logrus.Debug("platform buy_package resp body :", string(body))

			if err = json.Unmarshal(body, &buyResult); err == nil {
				if buyResult.ErrCode == 0 {
					order.Effective = 1 //流量包已经生效

					pOrder := &mysql.BuyPackagePlatform{
						UserId:   order.UserId,
						Uuid:     order.Uuid,
						DeviceSn: order.DeviceSn,

						PackageId: order.PackageId,
						Currency:  order.Currency,
						Count:     order.Count,
						Money:     order.Money,
						OrderTime: order.OrderTime,

						PlatformOrderId:     buyResult.OrderId,
						DevicePackageId:     buyResult.DevicePackageId,
						DevicePackageIdList: buyResult.DevicePackageIdList,
					}

					if err = mysql.UpdateOrderTX(order, pOrder); err == nil {
						if _, err = redis.DelKey(order.OrderId); err == nil {
							var res ProtocolRsp
							res.Code = OK
							res.Msg = SUCCESS
							res.Data = buyResult
							res.ResponseWriter(ctx)
							return
						}
					}
				} else {
					err = errors.New(buyResult.ErrDesc)
				}
			}
		}
	}

	var res ProtocolRsp
	res.Code = ReqPlatformErrCode
	if err != nil {
		res.Msg = err.Error()
	} else {
		res.Msg = PackagePlatformErr
	}
	res.ResponseWriter(ctx)

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
