package models

// Endpoint entity...
type Endpoint struct {
	IpAddress            string `json:"ipAddress"`
	ServerName           string `json:"serverName"`
	StatusMessage        string `json:"statusMessage"`
	Grade                string `json:"grade"`
	GradeTrustIgnored    string `json:"gradeTrustIgnored"`
	HasWarnings          bool   `json:"hasWarnings"`
	IsExceptional        bool   `json:"isExceptional"`
	Progress             int    `json:"progress"`
	Duration             int    `json:"duration"`
	StatusDetails        string `json:"statusDetails"`
	StatusDetailsMessage string `json:"statusDetailsMessage"`
	Delegation           int    `json:"delegation"`
}
