package models

import (
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
)

const (
	OK = 0

	NotLoginCode = 300 //未登录
	TokenExpCode = 301 //token失效

	LoginErrCode       = 303 //登录失败
	RegisterErrCode    = 304 //注册失败
	ReqPlatformErrCode = 305 //平台获取数据失败
	OrderErrCode       = 306 //生成订单号失败
	PayHistoryErrCode  = 307 //查询所有支付订单失败
	NewsErrCode        = 308 //查询新闻失败
	AccountErrCode     = 309 //查询账号失败

	ParamErrCode   = 401 //请求参数异常
	NoFoundErrCode = 404 //Not Found

	UnknownErrCode = 501 //未知异常

)

const (
	SUCCESS = "success"

	RegisterUserNameEmptyErr  = "Registered username is empty"
	RegisterUserNameFormatErr = "Registered username format error:Number or letter does not limit capitalization 6-30 digits"

	RegisterEmailEmptyErr  = "Registered email is empty"
	RegisterEmailFormatErr = "Registered email format error:Number or letter or symbol 6-30 digits"
	EmailEmptyErr          = "Email is empty"
	EmailFormatErr         = "Email format error:Number or letter or symbol 6-30 digits"

	RegisterMobileEmptyErr  = "Registered mobile is empty"
	RegisterMobileFormatErr = "Registered mobile format error"

	RegisterPassWordEmptyErr  = "Registered password is empty"
	RegisterPassWordFormatErr = "Registered password format error:Number or letter or symbol 6-30 digits"

	LoginErrUserNameOrEmailEmptyErr = "Login username or email is empty"
	LoginErrPassWordEmptyErr        = "Login password is empty"
	LoginErrPassWordFormatErr       = "Login password format error:Number + letter + symbol 6-30 digits"

	LoginUserNameFormatErr = "login username format error:Number or letter does not limit capitalization 6-30 digits"
	LoginEmailFormatErr    = "login email format error:Number or letter or symbol 6-30 digits"

	ParamErr           = "Request Param Error"
	GenOrderErr        = "Generate Order Error"
	PackageIdErr       = "Found Package Id Error"
	PackagePlatformErr = "Platform Package Error"
	TokenErr           = "Token Error"
	TokenExpiredErr    = "Token Expired"
	NoFoundErr         = "NoFound Error"
	UnknownErr         = "Unknown Error"
)

type ProtocolRsp struct {
	Data interface{} `json:"data,omitempty"`
}

type ErrorRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (json *ProtocolRsp) ResponseWriter(ctx iris.Context, status int) {

	ctx.StatusCode(status)

	if _, err := ctx.JSON(json); err != nil {
		logrus.Error(err)
	}
}

func (json *ErrorRsp) ResponseWriter(ctx iris.Context, status int) {

	ctx.StatusCode(status)

	if _, err := ctx.JSON(json); err != nil {
		logrus.Error(err)
	}
}
