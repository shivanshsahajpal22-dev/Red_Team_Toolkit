package model

type Target struct {
	Platform    string `json:"platform"`
	PlatformURL string `json:"platform_url"`
	Identifier  string `json:"identifier"`
	Type        string `json:"type"`
}

type ResultStatus string

const (
	StatusExists       ResultStatus = "exists"
	StatusNotFound     ResultStatus = "not_found"
	StatusAuthRequired ResultStatus = "auth_required"
	StatusChallenge    ResultStatus = "challenge"
	StatusRateLimited  ResultStatus = "rate_limited"
	StatusUnknown      ResultStatus = "unknown"
)

type Profile struct {
	Platform    string       `json:"platform"`
	Identifier  string       `json:"identifier"`
	Status      ResultStatus `json:"status"`
	DisplayName string       `json:"display_name"`
	Bio         string       `json:"bio"`
	Location    string       `json:"location"`
	Links       []string     `json:"links"`
	URL         string       `json:"url"`
}
