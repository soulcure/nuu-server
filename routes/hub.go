package routes

import (
	"fmt"
	"net/http"
	"nuu-server/models"
	"nuu-server/mysql"
	"nuu-server/redis"
	"nuu-server/utils"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris/v12"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	SecretKey = "nuu-server"
)

// 所有的路由
func Hub(app *iris.Application) {
	app.OnErrorCode(iris.StatusNotFound, notFound)
	app.OnErrorCode(iris.StatusInternalServerError, internalServerError)

	// register our routes.
	app.Get("/test", test)
	app.Post("/register", registerHandler)
	app.Post("/login", loginHandler)

	app.Post("/api/detail", models.PackageDetailToday)
	app.Post("/api/period", models.UsedDetailPeriod)
	app.Post("/api/sale", models.QueryPackageForSale)
	app.Post("/api/package", models.PackageQuery)
	app.Post("/api/setting", models.SettWifiPassword)
	app.Post("/api/forget/password", models.SendPasswordEmail)

	app.Get("/api/news", news)

	//need login
	app.Post("/api/update", tokenHandler, updateProfile)
	app.Post("/api/order", tokenHandler, genOrder)
	app.Post("/api/pay", tokenHandler, models.OrderPay)
	app.Post("/api/pay/history", tokenHandler, payHistory)
}

func test(ctx iris.Context) {
	models.PackageDetailToday(ctx)
}

func notFound(ctx iris.Context) {
	var res models.ErrorRsp
	res.Code = models.NoFoundErrCode
	res.Message = models.NoFoundErr
	res.ResponseWriter(ctx, http.StatusNotFound)
}

//当出现错误的时候，再试一次
func internalServerError(ctx iris.Context) {
	var res models.ErrorRsp
	res.Code = models.UnknownErrCode
	res.Message = models.UnknownErr
	res.ResponseWriter(ctx, http.StatusInternalServerError)
}

//用户注册处理函数
func registerHandler(ctx iris.Context) {
	username := ctx.FormValue("username")
	email := ctx.FormValue("email")
	mobile := ctx.FormValue("mobile")
	iso := ctx.FormValue("iso")
	password := ctx.FormValue("password")

	if checkRegisterFormat(ctx, username, email, mobile, iso, password) {
		userUuid := uuid.NewV4().String()
		logrus.Debug("user register uuid:", userUuid)
		if id, err := mysql.RegisterInsert(userUuid, username, email, mobile, iso, password); err == nil {
			logrus.Debug("user register success")
			userId := int(id)
			exp := time.Now().Add(time.Hour * 72).Unix()
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":   userId,
				"uuid": userUuid,
				"exp":  exp,
			})
			if token, err := token.SignedString([]byte(SecretKey)); err == nil {
				logrus.Debug(username, "  set Token:", token)
				var res models.ProtocolRsp

				res.Data = &models.RegisterRsp{Id: userId, Uuid: userUuid, UserName: username, Email: email, PassWord: password, Token: token, Expired: exp}
				res.ResponseWriter(ctx, http.StatusOK)
			} else {
				var res models.ErrorRsp
				res.Code = models.RegisterErrCode
				res.Message = err.Error()
				res.ResponseWriter(ctx, http.StatusBadRequest)
			}
		} else {
			var res models.ErrorRsp
			res.Code = models.RegisterErrCode
			res.Message = err.Error()

			res.ResponseWriter(ctx, http.StatusBadRequest)
		}

	}

}

func loginHandler(ctx iris.Context) {
	username := ctx.FormValue("username")
	email := ctx.FormValue("email")
	password := ctx.FormValue("password")

	if checkLoginFormat(ctx, username, email, password) {
		if account, err := mysql.AccountLogin(username, email, password); err == nil {

			exp := time.Now().Add(time.Hour * 72).Unix()
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":   account.Id,
				"uuid": account.Uuid,
				"exp":  exp,
			})

			if token, err := token.SignedString([]byte(SecretKey)); err == nil {
				logrus.Debug(username, "  set Token:", token)
				var res models.ProtocolRsp
				res.Data = &models.LoginRsp{Token: token, Expired: exp, Id: account.Id, Uuid: account.Uuid, UserName: account.UserName, Email: account.Email}
				res.ResponseWriter(ctx, http.StatusOK)
			} else {
				var res models.ErrorRsp
				res.Code = models.LoginErrCode
				res.Message = err.Error()
				res.ResponseWriter(ctx, http.StatusBadRequest)
			}

		} else {
			var res models.ErrorRsp
			res.Code = models.LoginErrCode
			res.Message = err.Error()
			res.ResponseWriter(ctx, http.StatusBadRequest)
		}
	}
}

func tokenHandler(ctx iris.Context) {
	tokenString := ctx.GetHeader("token")
	if tokenString == "" {
		//ctx.StatusCode(http.StatusUnauthorized)
		var res models.ErrorRsp
		res.Code = models.NotLoginCode
		res.Message = models.TokenErr
		res.ResponseWriter(ctx, http.StatusBadRequest)

		logrus.Error("Unauthorized access to this resource")
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err == nil && token.Valid {
		logrus.Debug("Token is valid")

		Claims := token.Claims
		ctx.Values().Set("id", Claims.(jwt.MapClaims)["id"])
		ctx.Values().Set("uuid", Claims.(jwt.MapClaims)["uuid"])

		ctx.Next()
	} else {
		//ctx.StatusCode(http.StatusUnauthorized)
		var res models.ErrorRsp
		res.Code = models.TokenExpCode
		res.Message = models.TokenExpiredErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		logrus.Error("Token is invalid!!!")
	}

}

func updateProfile(ctx iris.Context) {
	userId := ctx.FormValue("userId")
	email := ctx.FormValue("email")
	genderStr := ctx.FormValue("gender")
	gender, err := strconv.Atoi(genderStr)
	if err != nil {
		gender = 0
		logrus.Error(err)
	}

	if userId != "" && email != "" {
		if err := mysql.Update(userId, gender, email); err == nil {
			logrus.Debug("user update profile success")
			var e error

			user := &mysql.Account{}
			e = redis.GetStruct(userId, user)

			logrus.Debug("user profile:", user)

			user.Email = email
			_, e = redis.SetStruct(userId, user)

			if e == nil {
				var res models.ProtocolRsp
				res.Data = user
				res.ResponseWriter(ctx, http.StatusOK)
				return
			}
		}

	}
}

func genOrder(ctx iris.Context) {
	orderId := redis.GetOrderNum()
	orderTime := time.Now().Format("20060102150405")

	deviceSn := ctx.FormValue("deviceSn")
	price := ctx.FormValue("price") //need server check price
	currency := ctx.FormValue("currency")
	beginDate := ctx.FormValue("beginDate")
	packageName := ctx.FormValue("packageName")

	packageId, err := ctx.PostValueInt("packageId")
	if err != nil {
		var res models.ErrorRsp
		res.Code = models.ParamErrCode
		res.Message = models.PackageIdErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return
	}

	count, err := ctx.PostValueInt("count")
	if err != nil {
		count = 1
	}

	effectiveType, err := ctx.PostValueInt("effective_type")
	if err != nil {
		effectiveType = 0
	}

	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		var res models.ErrorRsp
		res.Code = models.ParamErrCode
		res.Message = models.ParamErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return
	}

	money := p * float64(count)
	moneyStr := fmt.Sprintf("%.2f", money)

	order := mysql.OrderReq{
		UserId:      int(ctx.Values().Get("id").(float64)),
		Uuid:        ctx.Values().Get("uuid").(string),
		OrderId:     orderId,
		Price:       price,
		Currency:    currency,
		DeviceSn:    deviceSn,
		PackageId:   packageId,
		PackageName: packageName,
		OrderTime:   orderTime,
		BeginDate:   beginDate,

		Status: 0,  //0未支付
		PayId:  "", //等待客户端上传payment id
		Count:  uint8(count),
		Money:  moneyStr,

		Effective:     0,                    //0流量包未生效
		EffectiveType: uint8(effectiveType), //生效类型
		Discount:      100,                  //商品未打折
	}

	if id, err := order.InsertOrder(); err == nil {
		order.Id = int(id)
		if _, err = redis.SetStruct(orderId, order); err == nil {
			var res models.ProtocolRsp
			res.Data = &mysql.OrderRsp{OrderId: orderId, Money: moneyStr}
			res.ResponseWriter(ctx, http.StatusOK)
			return
		}
	}

	var res models.ErrorRsp
	res.Code = models.OrderErrCode
	res.Message = models.GenOrderErr
	res.ResponseWriter(ctx, http.StatusBadRequest)

}

//查询所有支付的订单
func payHistory(ctx iris.Context) {
	userId := int(ctx.Values().Get("id").(float64))

	if pays, err := mysql.QueryPayHistory(userId); err == nil {
		var res models.ProtocolRsp
		res.Data = pays
		res.ResponseWriter(ctx, http.StatusOK)
	} else {
		var res models.ErrorRsp
		res.Code = models.PayHistoryErrCode
		res.Message = err.Error()
		res.ResponseWriter(ctx, http.StatusBadRequest)
	}

}

//查询所有新闻
func news(ctx iris.Context) {
	if pays, err := mysql.News(); err == nil {
		var res models.ProtocolRsp
		res.Data = pays
		res.ResponseWriter(ctx, http.StatusOK)
	} else {
		var res models.ErrorRsp
		res.Code = models.NewsErrCode
		res.Message = err.Error()
		res.ResponseWriter(ctx, http.StatusBadRequest)
	}
}

func checkRegisterFormat(ctx iris.Context, username, email, mobile, iso, password string) bool {
	if username == "" {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterUserNameEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	} else if !utils.IsUserName(username) {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterUserNameFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	}
	if email == "" {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterEmailEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	} else if !utils.IsEmail(email) {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterEmailFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	}

	if mobile == "" || iso == "" {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterMobileEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	} else if !utils.IsMobile(mobile, iso) {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterMobileFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	}

	if password == "" {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterPassWordEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	} else if !utils.IsPwd(password) {
		var res models.ErrorRsp
		res.Code = models.RegisterErrCode
		res.Message = models.RegisterPassWordFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	}

	return true
}

func checkLoginFormat(ctx iris.Context, username, email, password string) bool {
	if username == "" && email == "" {
		var res models.ErrorRsp
		res.Code = models.LoginErrCode
		res.Message = models.LoginErrUserNameOrEmailEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	} else if username == "" && !utils.IsEmail(email) {
		var res models.ErrorRsp
		res.Code = models.LoginErrCode
		res.Message = models.LoginEmailFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	} else if email == "" && !utils.IsUserName(username) {
		var res models.ErrorRsp
		res.Code = models.LoginErrCode
		res.Message = models.LoginUserNameFormatErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	}

	if password == "" {
		var res models.ErrorRsp
		res.Code = models.LoginErrCode
		res.Message = models.LoginErrPassWordEmptyErr
		res.ResponseWriter(ctx, http.StatusBadRequest)
		return false
	}

	return true
}
