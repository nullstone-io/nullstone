package gcp

type ServiceAccount struct {
	Email      string `json:"email"`
	PrivateKey string `json:"private_key"`
}
