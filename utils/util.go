package utils

import (
	"regexp"

	"github.com/nyaruka/phonenumbers"
)

/************************* 自定义类型 ************************/ //IsUserName 数字+字母  不限制大小写 6~30位

//IsUserName 数字+字母  不限制大小写 6~30位
func IsUserName(str string) bool {
	b, _ := regexp.MatchString("^[0-9a-zA-Z]{6,30}$", str)
	return b
}

//IsPwd 数字+字母+符号 6~30位
func IsPwd(str string) bool {
	b, _ := regexp.MatchString("^[0-9a-zA-Z@.]{6,30}$", str)
	return b
}

//IsEmail 邮箱 最高30位
func IsEmail(str string) bool {
	b, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", str)
	return b
}

//IsMobile 手机号码 带国家码号码
func IsMobile(str, iso string) bool {
	// parse our phone number
	if num, err := phonenumbers.Parse(str, iso); err == nil {
		return phonenumbers.IsValidNumber(num)
	}
	return false

}
