package mysql

type PayPalAuth struct {
	Scope       string `json:"scope"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	AppId       string `json:"app_id"`
	ExpiresIn   string `json:"expires_in"`
	Nonce       string `json:"nonce"`
}

// Link struct
type Link struct {
	Href    string `json:"href"`
	Rel     string `json:"rel,omitempty"`
	Method  string `json:"method,omitempty"`
	Enctype string `json:"enctype,omitempty"`
}
