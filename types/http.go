package types

type SSHData struct {
	SecretKey string `json:"secret_key"`
	PublicKey string `json:"public_key"`
	Passwd    string `json:"passwd"`
}
