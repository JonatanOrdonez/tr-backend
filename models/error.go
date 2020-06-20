package models

// Error entity...
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
