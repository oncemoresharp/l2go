package models

type Account struct {
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	AccessLevel int8   `json:"access_level"`
}
