package types

type GetAllSecretsRequest struct {
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

type GetReportRequest struct {
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	SecretID  string `json:"secret_id"`
}
