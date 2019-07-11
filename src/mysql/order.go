package mysql

import (
	"database/sql"
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
	PackageName   string `db:"package_name" json:"packageName"` //流量包名称
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

	PackageId   int    `db:"package_id" json:"packageId"`
	PackageName string `db:"package_name" json:"packageName"` //流量包名称
	Currency    string `db:"currency" json:"currency"`        //货币 CNY,USD,HKD
	Count       uint8  `db:"count" json:"count"`              //流量包数量
	Money       string `db:"money" json:"money"`              //支付总金额
	OrderTime   string `db:"order_time" json:"orderTime"`

	PlatformOrderId     string `db:"platform_order_id" redis:"platform_order_id"`
	DevicePackageId     int    `db:"device_package_id" redis:"device_package_id"`
	DevicePackageIdList string `db:"device_package_id_list" redis:"device_package_id_list"`
}

type NewsBean struct {
	Id      int    `db:"id" redis:"id,omitempty"`
	Title   string `db:"title" json:"title,omitempty"`
	Time    string `db:"time" redis:"time,omitempty"`
	Content string `db:"content" redis:"content,omitempty"`
}

func (order *OrderReq) InsertOrder() (int64, error) {
	r, err := db.Exec("insert into c_order(user_id,uuid,order_id,price,currency,device_sn,package_id,package_name,order_time,begin_date,status,pay_id,count,money,effective,effective_type,discount)values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		order.UserId, order.Uuid, order.OrderId, order.Price, order.Currency, order.DeviceSn, order.PackageId, order.PackageName, order.OrderTime, order.BeginDate, order.Status, order.PayId, order.Count, order.Money, order.Effective, order.EffectiveType, order.Discount)
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
	r, err := db.Exec("insert into p_order(user_id,uuid,device_sn,package_id,package_name,currency,count,money,order_time,platform_order_id,device_package_id,device_package_id_list)values(?,?,?,?,?,?,?,?,?,?,?,?)",
		order.UserId, order.Uuid, order.DeviceSn, order.PackageId, order.PackageName, order.Currency, order.Count, order.Money, order.OrderTime, order.PlatformOrderId, order.DevicePackageId, order.DevicePackageIdList)
	if err != nil {
		logrus.Error("mysql Insert order err", err)
		return 0, err
	}
	logrus.Info("mysql Insert order success :%v", r)
	return r.LastInsertId()
}

//查询账号的所有支付历史
func QueryPayHistory(userId int) ([]BuyPackagePlatform, error) {
	var (
		res  []BuyPackagePlatform
		rows *sql.Rows
		err  error
	)

	rows, err = db.Query("select * from p_order where user_id = ?", userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item BuyPackagePlatform
		if err = rows.Scan(&item.Id, &item.UserId, &item.Uuid, &item.DeviceSn, &item.PackageId, &item.PackageName, &item.Currency, &item.Count, &item.Money, &item.OrderTime, &item.PlatformOrderId, &item.DevicePackageId, &item.DevicePackageIdList); err == nil {
			res = append([]BuyPackagePlatform{item}, res...) // 在开头添加1个元素
		} else {
			break
		}
	}
	return res, err
}

//查询账号的所有支付历史
func News() ([]NewsBean, error) {
	var (
		res  []NewsBean
		rows *sql.Rows
		err  error
	)

	rows, err = db.Query("select * from news")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item NewsBean
		if err = rows.Scan(&item.Id, &item.Title, &item.Title, &item.Content); err == nil {
			res = append([]NewsBean{item}, res...) // 在开头添加1个元素
		} else {
			break
		}
	}
	return res, err
}

func UpdateOrderTX(order *OrderReq, pOrder *BuyPackagePlatform) error {
	tx, e := db.Begin()
	if e != nil {
		return e
	}

	r, err := tx.Exec("update c_order set status = ?,pay_id = ?,effective = ? where order_id = ? and status =0",
		order.Status, order.PayId, order.Effective, order.OrderId)
	if err != nil {
		logrus.Error("mysql update order status err :", err)
		txRollback(tx)
		return err
	} else {
		rowAffected, err := r.RowsAffected()
		if err != nil {
			txRollback(tx)
			logrus.Error("mysql update c_order error:", rowAffected)
			return err
		} else {
			logrus.Debug("mysql update c_order success:", rowAffected)
		}
	}
	logrus.Info("mysql update order status success", order.OrderId)

	r, err = tx.Exec("insert into p_order(user_id,uuid,device_sn,package_id,package_name,currency,count,money,order_time,platform_order_id,device_package_id,device_package_id_list)values(?,?,?,?,?,?,?,?,?,?,?,?)",
		pOrder.UserId, pOrder.Uuid, pOrder.DeviceSn, pOrder.PackageId, pOrder.PackageName, pOrder.Currency, pOrder.Count, pOrder.Money, pOrder.OrderTime, pOrder.PlatformOrderId, pOrder.DevicePackageId, pOrder.DevicePackageIdList)
	if err != nil {
		txRollback(tx)
		logrus.Error("mysql insert p_order error:", err)
		return err
	} else {
		rowAffected, err := r.RowsAffected()
		if err != nil {
			txRollback(tx)
			logrus.Error("mysql insert p_order error:", rowAffected)
			return err
		} else {
			logrus.Debug("mysql insert p_order success:", rowAffected)
		}
	}

	err = tx.Commit()
	if err != nil {
		txRollback(tx)
		return err
	}
	return err
}

func txRollback(tx *sql.Tx) {
	e := tx.Rollback()
	if e != nil {
		logrus.Error("tx.Rollback() Error:" + e.Error())
	}
}
