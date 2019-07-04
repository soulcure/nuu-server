package routes

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"models"
	"mysql"
	"net/http"
	"redis"
	"strconv"
	"time"
	"utils"
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
	app.Post("/api/update", tokenHandler, updateProfile)
}

func test(ctx iris.Context) {
	models.PackageDetailToday(ctx, "354243074362656")
}

func notFound(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotFound)
	var res models.ProtocolRsp
	res.Code = models.NoFoundErrCode
	res.Msg = models.NoFoundErr
	res.ResponseWriter(ctx)
}

//当出现错误的时候，再试一次
func internalServerError(ctx iris.Context) {
	ctx.StatusCode(http.StatusRequestTimeout)
	var res models.ProtocolRsp
	res.Code = models.UnknownErrCode
	res.Msg = models.UnknownErr
	res.ResponseWriter(ctx)
}

//用户注册处理函数
func registerHandler(ctx iris.Context) {
	username := ctx.FormValue("username")
	email := ctx.FormValue("email")
	mobile := ctx.FormValue("mobile")
	iso := ctx.FormValue("iso")
	password := ctx.FormValue("password")

	if checkRegisterFormat(ctx, username, email, mobile, iso, password) {
		userUuid := uuid.Must(uuid.NewV4()).String()
		logrus.Debug("user register uuid:", userUuid)
		if id, err := mysql.RegisterInsert(userUuid, username, email, mobile, iso, password); err == nil {
			logrus.Debug("user register success")
			var res models.ProtocolRsp
			res.Code = models.OK
			res.Msg = models.SUCCESS
			res.Data = &models.RegisterRsp{Id: int(id), Uuid: userUuid, UserName: username, Email: email, PassWord: password}
			res.ResponseWriter(ctx)

		} else {
			var res models.ProtocolRsp
			res.Code = models.RegisterErrCode
			res.Msg = err.Error()
			res.ResponseWriter(ctx)
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
				res.Code = models.OK
				res.Msg = models.SUCCESS
				res.Data = &models.LoginRsp{Token: token, Id: account.Id, Uuid: account.Uuid, Expired: exp}
				res.ResponseWriter(ctx)
			} else {
				var res models.ProtocolRsp
				res.Code = models.LoginErrCode
				res.Msg = err.Error()
				res.ResponseWriter(ctx)
			}

		}
	}
}

func tokenHandler(ctx iris.Context) {
	tokenString := ctx.GetHeader("token")
	if tokenString == "" {
		ctx.StatusCode(http.StatusUnauthorized)
		var res models.ProtocolRsp
		res.Code = models.NotLoginCode
		res.Msg = models.TokenErr
		res.ResponseWriter(ctx)

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
		ctx.StatusCode(http.StatusUnauthorized)
		var res models.ProtocolRsp
		res.Code = models.TokenExpCode
		res.Msg = models.TokenExpiredErr
		res.ResponseWriter(ctx)
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
				res.Code = models.OK
				res.Msg = models.SUCCESS
				res.ResponseWriter(ctx)
				return
			}
		}

	}
}

func checkRegisterFormat(ctx iris.Context, username, email, mobile, iso, password string) bool {
	if username == "" {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterUserNameEmptyErr
		res.ResponseWriter(ctx)
		return false
	} else if !utils.IsUserName(username) {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterUserNameFormatErr
		res.ResponseWriter(ctx)
		return false
	}
	if email == "" {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterEmailEmptyErr
		res.ResponseWriter(ctx)
		return false
	} else if !utils.IsEmail(email) {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterEmailFormatErr
		res.ResponseWriter(ctx)
		return false
	}

	if mobile == "" || iso == "" {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterMobileEmptyErr
		res.ResponseWriter(ctx)
		return false
	} else if !utils.IsMobile(mobile, iso) {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterMobileFormatErr
		res.ResponseWriter(ctx)
		return false
	}

	if password == "" {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterPassWordEmptyErr
		res.ResponseWriter(ctx)
		return false
	} else if !utils.IsPwd(password) {
		var res models.ProtocolRsp
		res.Code = models.RegisterErrCode
		res.Msg = models.RegisterPassWordFormatErr
		res.ResponseWriter(ctx)
		return false
	}

	return true
}

func checkLoginFormat(ctx iris.Context, username, email, password string) bool {
	if username == "" && email == "" {
		var res models.ProtocolRsp
		res.Code = models.LoginErrCode
		res.Msg = models.LoginErrUserNameOrEmailEmptyErr
		res.ResponseWriter(ctx)
		return false
	} else if username == "" && !utils.IsEmail(email) {
		var res models.ProtocolRsp
		res.Code = models.LoginErrCode
		res.Msg = models.LoginEmailFormatErr
		res.ResponseWriter(ctx)
		return false
	} else if email == "" && !utils.IsUserName(username) {
		var res models.ProtocolRsp
		res.Code = models.LoginErrCode
		res.Msg = models.LoginUserNameFormatErr
		res.ResponseWriter(ctx)
		return false
	}

	if password == "" {
		var res models.ProtocolRsp
		res.Code = models.LoginErrCode
		res.Msg = models.LoginErrPassWordEmptyErr
		res.ResponseWriter(ctx)
		return false
	}

	return true
}
