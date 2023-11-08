package types

type FromGetAllSecretsMiddlewareToHandler struct {
	FoundedAccessLog map[string][]AccessLog
	FoundedSecrets   map[string]Secret
	ArnList          []string
	FoundedArnList   bool
	PublicKey        string
	SecretKey        string
	Region           string
}

type FromGetReportMiddlewareToHandler struct {
	FoundedSecret    *Secret
	FoundedAccessLog []AccessLog
	PublicKey        string
	SecretKey        string
	SecretID         string
	Region           string
}
