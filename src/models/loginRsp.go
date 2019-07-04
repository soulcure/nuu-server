package models

type LoginRsp struct {
	Token   string `json:"token"`
	Id      int    `json:"id"`
	Uuid    string `json:"uuid"`
	Expired int64  `json:"expired"`
}
