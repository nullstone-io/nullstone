package aws

// User contains credentials for a user that has access to perform a particular action in AWS
// This structure must match the fields defined in outputs of the module
type User struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key"`
	SecretAccessKey string `json:"secret_key"`
}
