package types

type GetAllSecretsResponse struct {
	Secrets   []Secret               `json:"secrets"`
	AccessLog map[string][]AccessLog `json:"access_logs"`
	Error     string                 `json:"error"`
}

type GetReportResponse struct {
	Report string `json:"report"`
	Error  string `json:"error"`
}
