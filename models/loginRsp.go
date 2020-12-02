package models

// LoginRsp 登录结果结构体
// swagger:model LoginRsp
type LoginRsp struct {
	// discriminator: true
	// swagger:name token
	Token string `json:"token"`

	// discriminator: true
	// swagger:name expired
	Expired int64 `json:"expired"`

	// discriminator: true
	// swagger:name id
	ID int `json:"id"`

	// discriminator: true
	// swagger:name uuid
	UUID string `json:"uuid"`

	// discriminator: true
	// swagger:name username
	UserName string `json:"username"`

	// discriminator: true
	// swagger:name email
	Email string `json:"email"`
}
