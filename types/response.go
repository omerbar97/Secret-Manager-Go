package types

type GetAllSecretsResponse struct {
	Secrets   []Secret               `json:"secrets"`
	AccessLog map[string][]AccessLog `json:"access_logs"`
}

type GetReportResponse struct {
	Report string `json:"report"`
}
