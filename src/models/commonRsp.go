package models

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
)

const (
	OK = 0

	NotLoginCode       = 301 //未登录
	LoginErrCode       = 302 //登录失败
	TokenExpCode       = 303 //token失效
	RegisterErrCode    = 304 //注册失败
	ReqPlatformErrCode = 305 //平台获取数据失败
	OrderErrCode       = 306 //生成订单号失败

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

	RegisterMobileEmptyErr  = "Registered mobile is empty"
	RegisterMobileFormatErr = "Registered mobile format error"

	RegisterPassWordEmptyErr  = "Registered password is empty"
	RegisterPassWordFormatErr = "Registered password format error:Number or letter or symbol 6-30 digits"

	LoginErrUserNameOrEmailEmptyErr = "Login username or email is empty"
	LoginErrPassWordEmptyErr        = "Login password is empty"
	LoginErrPassWordFormatErr       = "Login password format error:Number + letter + symbol 6-30 digits"

	LoginUserNameFormatErr = "login username format error:Number or letter does not limit capitalization 6-30 digits"
	LoginEmailFormatErr    = "login email format error:Number or letter or symbol 6-30 digits"

	ParamErr        = "Request Param Error"
	GenOrderErr     = "Generate Order Error"
	PackageIdErr    = "Found Package Id Error"
	TokenErr        = "Token Error"
	TokenExpiredErr = "Token Expired"
	NoFoundErr      = "NoFound Error"
	UnknownErr      = "Unknown Error"
)

type ProtocolRsp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (json *ProtocolRsp) ResponseWriter(ctx iris.Context) {
	if _, err := ctx.JSON(json); err != nil {
		logrus.Error(err)
	}
}
