package models

type Secret struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	User      string `json:"user"`
	State     string `json:"state"`
	Ctime     string `json:"ctime"`
}