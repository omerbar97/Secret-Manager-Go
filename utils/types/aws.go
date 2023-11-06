package types

import "time"

// Holding the information of the secret
type Secret struct {
	Name         string
	ARN          string
	Version      string
	CreatedAt    time.Time
	LastAccessed time.Time
}

// Holding the information of the accesslog
type AccessLog struct {
	User        string
	EventName   string
	EventSource string
	EventTime   time.Time
}

// Struct to store only 1 secret with it's paring accesslog
type SingleSecretWithAccessLog struct {
	Secret    Secret
	AccessLog []AccessLog
}

// Struct that returns from the AllSecrets func inside the AWS API
type AllSecretWithAccessLog struct {
	Secrets   []Secret
	AccessLog map[string][]AccessLog
}

// Struct that returns from the client.go file inside AWS service
type AllAccessLog struct {
	AccessLog []AccessLog
	NextToken *string
}

// Struct that returns from the client.go file inside AWS service
type AllSecrets struct {
	Secrets   []Secret
	NextToken *string
}
