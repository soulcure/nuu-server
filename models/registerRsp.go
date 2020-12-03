package models

//RegisterRsp 注册结果
type RegisterRsp struct {
	ID       int    `json:"id"`
	UUID     string `json:"uuid"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	PassWord string `json:"password"`

	Token   string `json:"token"`
	Expired int64  `json:"expired"`
}
