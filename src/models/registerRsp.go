package models

type RegisterRsp struct {
	Id       int    `json:"id"`
	Uuid     string `json:"uuid"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	PassWord string `json:"password"`
}
