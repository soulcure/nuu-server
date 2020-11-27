package models

type LoginRsp struct {
	Token    string `json:"token"`
	Expired  int64  `json:"expired"`
	Id       int    `json:"id"`
	Uuid     string `json:"uuid"`
	UserName string `json:"username"`
	Email    string `json:"email"`
}
