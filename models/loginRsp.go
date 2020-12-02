package models

// LoginRsp 登录结果结构体
type LoginRsp struct {
	Token    string `json:"token"`
	Expired  int64  `json:"expired"`
	ID       int    `json:"id"`
	UUID     string `json:"uuid"`
	UserName string `json:"username"`
	Email    string `json:"email"`
}
