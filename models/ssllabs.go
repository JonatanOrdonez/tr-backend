package models

// Ssllabs entity...
type Ssllabs struct {
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Protocol        string     `json:"protocol"`
	IsPublic        bool       `json:"isPublic"`
	Status          string     `json:"status"`
	StartTime       int64      `json:"startTime"`
	TestTime        int64      `json:"testTime"`
	EngineVersion   string     `json:"engineVersion"`
	CriteriaVersion string     `json:"criteriaVersion"`
	StatusMessage   string     `json:"statusMessage"`
	CacheExpiryTime int64      `json:"cacheExpiryTime"`
	Endpoints       []Endpoint `json:"endpoints"`
}
