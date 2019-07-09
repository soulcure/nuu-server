package mysql

import (
	"github.com/sirupsen/logrus"
)

type OrderRsp struct {
	OrderId string `json:"orderId"`
	Money   string `json:"money"`
}

type OrderReq struct {
	Id            int    `db:"id" json:"id"`
	UserId        int    `db:"user_id" json:"userId"`
	Uuid          string `db:"uuid" json:"uuid"`
	OrderId       string `db:"order_id" json:"orderId"`
	Price         string `db:"price" json:"price"`       //10.00
	Currency      string `db:"currency" json:"currency"` //货币 CNY,USD,HKD
	DeviceSn      string `db:"device_sn" json:"deviceSN"`
	PackageId     int    `db:"package_id" json:"packageId"`
	OrderTime     string `db:"order_time" json:"orderTime"`
	BeginDate     string `db:"begin_date" json:"begin_date"`        //流量包生效日期
	Status        uint8  `db:"status" json:"status"`                //支付状态  0未支付，1已经支付
	PayId         string `db:"pay_id" json:"payId"`                 //支付平台 支付ID
	Count         uint8  `db:"count" json:"count"`                  //流量包数量
	Money         string `db:"money" json:"money"`                  //支付总金额
	Effective     uint8  `db:"effective" json:"effective"`          //流量包是否生效，通知管理平台生效 0未生效，1已经生效
	EffectiveType uint8  `db:"effective_type" json:"EffectiveType"` //生效类型
	Discount      uint8  `db:"discount" json:"discount"`            //折扣  (0-100)
}

type BuyPackagePlatform struct {
	Id       int    `db:"id" redis:"id,omitempty"`
	UserId   int    `db:"user_id" json:"userId"`
	Uuid     string `db:"uuid" redis:"uuid"`
	DeviceSn string `db:"device_sn" redis:"device_sn"`

	PackageId int    `db:"package_id" json:"packageId"`
	Currency  string `db:"currency" json:"currency"` //货币 CNY,USD,HKD
	Count     uint8  `db:"count" json:"count"`       //流量包数量
	Money     string `db:"money" json:"money"`       //支付总金额
	OrderTime string `db:"order_time" json:"orderTime"`

	PlatformOrderId     string `db:"platform_order_id" redis:"platform_order_id"`
	DevicePackageId     int    `db:"device_package_id" redis:"device_package_id"`
	DevicePackageIdList string `db:"device_package_id_list" redis:"device_package_id_list"`
}

func (order *OrderReq) InsertOrder() (int64, error) {
	r, err := db.Exec("insert into c_order(user_id,uuid,order_id,price,currency,device_sn,package_id,order_time,begin_date,status,pay_id,count,money,effective,effective_type,discount)values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		order.UserId, order.Uuid, order.OrderId, order.Price, order.Currency, order.DeviceSn, order.PackageId, order.OrderTime, order.BeginDate, order.Status, order.PayId, order.Count, order.Money, order.Effective, order.EffectiveType, order.Discount)
	if err != nil {
		logrus.Error("mysql Insert order err", err)
		return 0, err
	}
	logrus.Info("mysql Insert order success :%v", r)
	return r.LastInsertId()
}

func (order *OrderReq) UpdateOrderStatus() error {
	_, err := db.Exec("update c_order set status = ?,pay_id = ?,effective = ? where order_id = ? and status =0", order.Status, order.PayId, order.Effective, order.OrderId)
	if err != nil {
		logrus.Error("mysql update order status err :%v", err)
		return err
	}
	logrus.Info("mysql update order status success", order.OrderId)
	return err
}

func (order *BuyPackagePlatform) InsertPlatformOrder() (int64, error) {
	r, err := db.Exec("insert into p_order(user_id,uuid,device_sn,package_id,currency,count,money,order_time,platform_order_id,device_package_id,device_package_id_list)values(?,?,?,?,?,?,?,?,?,?,?)",
		order.UserId, order.Uuid, order.DeviceSn, order.PackageId, order.Currency, order.Count, order.Money, order.OrderTime, order.PlatformOrderId, order.DevicePackageId, order.DevicePackageIdList)
	if err != nil {
		logrus.Error("mysql Insert order err", err)
		return 0, err
	}
	logrus.Info("mysql Insert order success :%v", r)
	return r.LastInsertId()
}
