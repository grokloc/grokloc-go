package app

// Token describes the token value and the expiration unixtime
type Token struct {
	Bearer  string `json:"bearer"`
	Expires int64  `json:"expires"`
}
