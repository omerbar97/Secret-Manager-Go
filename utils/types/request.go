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

// When passing the value of the context to another handler/middleware
// will use this string
type contextKey string

var toContextKey contextKey = "contextFromHandler"

func GetContextInforamtionKey() contextKey {
	return toContextKey
}
