package models

// Domain entity...
type Domain struct {
	Servers          []Server `db:"servers" json:"servers"`
	ServersChanged   bool     `json:"servers_changed"`
	SslGrade         string   `db:"url" json:"ssl_grade"`
	PreviousSslGrade string   `db:"previousSslGrade" json:"previous_ssl_grade"`
	Logo             string   `db:"logo" json:"logo"`
	Title            string   `db:"title" json:"title"`
	IsDown           bool     `json:"is_down"`
	Id               int64    `db:"id" json:"-"`
	Url              string   `db:"url" json:"url"`
	UpdatedAt        int64    `db:"updatedAt" json:"-"`
}
