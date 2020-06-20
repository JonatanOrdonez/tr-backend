package repositories

import (
	"database/sql"
	"encoding/json"

	models "github.com/JonatanOrdonez/tr-backend/models"
)

// DomainRepo: Structure used to store the database access reference
type DomainRepo struct {
	db *sql.DB
}

// NewDomainRepository: Receives a reference to the database and stores it in the DomainRepo structure
// Params:
// (db): Reference to the sql.DB database object
// Return:
// (*DomainRepo): Reference to the DomainRepo object
func NewDomainRepository(db *sql.DB) *DomainRepo {
	return &DomainRepo{db: db}
}

// FindByID: Searchs for a domain in the database using its id property as a search criteria
// Params:
// (ID): Id of the domain you are looking for
// Return:
// (*models.Domain): Reference to the domain that was found
// (error): Error if the process fails
func (r *DomainRepo) FindByID(ID int64) (*models.Domain, error) {
	var id, updatedAt int64
	var url, sslGrade, previousSslGrade, logo, title string
	var servers []byte
	err := r.db.QueryRow("SELECT * FROM domains WHERE id=$1", ID).Scan(&id, &servers, &url, &sslGrade, &previousSslGrade, &logo, &title, &updatedAt)
	if err != nil {
		return nil, err
	}
	serversStruct, decodeErr := jsonToServers(servers)
	if decodeErr != nil {
		return nil, decodeErr
	}
	domain := &models.Domain{serversStruct, false, sslGrade, previousSslGrade, logo, title, false, id, url, updatedAt}
	return domain, nil
}

// GetAll: Gets all the records that are in the "domains" table
// Return:
// ([]*models.Domain): reference to the domain slice
// (error): Error if the process fails
func (r *DomainRepo) GetAll() ([]*models.Domain, error) {
	rows, err := r.db.Query("SELECT * FROM domains")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	domains := make([]*models.Domain, 0)
	for rows.Next() {
		var id, updatedAt int64
		var url, sslGrade, previousSslGrade, logo, title string
		var servers []byte
		err := rows.Scan(&id, &servers, &url, &sslGrade, &previousSslGrade, &logo, &title, &updatedAt)
		if err != nil {
			return nil, err
		}
		serversStruct, decodeErr := jsonToServers(servers)
		if decodeErr != nil {
			return nil, decodeErr
		}
		domain := &models.Domain{serversStruct, false, sslGrade, previousSslGrade, logo, title, false, id, url, updatedAt}
		domains = append(domains, domain)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return domains, nil
}

// Save: Stores a new domain in the database
// Params:
// (domain): Reference to the domain object to be stored
// Return:
// (int64): Id of the stored domain
// (error): Error if the process fails
func (r *DomainRepo) Save(domain *models.Domain) (int64, error) {
	id := int64(-1)
	jsonServers, jsonError := serversToJson(domain.Servers)
	if jsonError != nil {
		return id, jsonError
	}
	queryErr := r.db.QueryRow(`INSERT INTO domains (servers, url, sslGrade, previousSslGrade, logo, title, updatedAt) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, jsonServers, domain.Url, domain.SslGrade, domain.PreviousSslGrade, domain.Logo, domain.Title, domain.UpdatedAt).Scan(&id)
	if queryErr != nil {
		return id, queryErr
	}
	return id, nil
}

// Update: Update a domain in the database
// Params:
// (domain): Reference to the domain object to be updated
// Return:
// (int64): Id of the updated domain
// (error): Error if the process fails
func (r *DomainRepo) Update(domain *models.Domain) (int64, error) {
	id := int64(-1)
	jsonServers, jsonError := serversToJson(domain.Servers)
	if jsonError != nil {
		return id, jsonError
	}
	_, queryErr := r.db.Exec(`UPDATE domains SET servers=$1, url=$2, sslGrade=$3, previousSslGrade=$4, logo=$5, title=$6, updatedAt=$7 WEHRE id=$8`, jsonServers, domain.Url, domain.SslGrade, domain.PreviousSslGrade, domain.Logo, domain.Title, domain.UpdatedAt, domain.Id)
	if queryErr != nil {
		return id, queryErr
	}
	id = domain.Id
	return id, nil
}

// Delete: Remove a record from the domain table
// Params:
// (ID): Id of the domain you want to remove
// Return:
// (error): Error if the process fails
func (r *DomainRepo) Delete(ID int) error {
	return nil
}

// FindByUrl: Searchs for a domain in the database using its query property as a search criteria
// Params:
// (Url): URl of the domain you are looking for
// Return:
// (*models.Domain): Reference to the domain that was found
// (error): Error if the process fails
func (r *DomainRepo) FindByUrl(Url string) (*models.Domain, error) {
	var id, updatedAt int64
	var url, sslGrade, previousSslGrade, logo, title string
	var servers []byte
	err := r.db.QueryRow("SELECT * FROM domains WHERE url=$1", Url).Scan(&id, &servers, &url, &sslGrade, &previousSslGrade, &logo, &title, &updatedAt)
	if err != nil {
		return nil, err
	}
	serversStruct, decodeErr := jsonToServers(servers)
	if decodeErr != nil {
		return nil, decodeErr
	}
	domain := &models.Domain{serversStruct, false, sslGrade, previousSslGrade, logo, title, false, id, url, updatedAt}
	return domain, nil
}

// serversToJson: Auxiliary function to parse a server-type object slice to JSON format
// Params:
// (servers): Domain slice
// Return:
// ([]byte): JSON object
// (error): Error if the process fails
func serversToJson(servers []models.Server) ([]byte, error) {
	return json.Marshal(servers)
}

// jsonToServers: Auxiliary function to parse a JSON format to server-type object slice
// Params:
// (jsonData): JSON format
// Return:
// ([]models.Server): Domain slice
// (error): Error if the process fails
func jsonToServers(jsonData []byte) ([]models.Server, error) {
	var servers []models.Server
	err := json.Unmarshal(jsonData, &servers)
	if err != nil {
		return nil, err
	}
	return servers, nil
}
