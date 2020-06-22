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
	var servers, endpoints []byte
	var serversChanged, isDown bool
	err := r.db.QueryRow("SELECT * FROM domains WHERE id=$1", ID).Scan(&id, &servers, &endpoints, &url, &sslGrade, &previousSslGrade, &logo, &title, &updatedAt, &serversChanged, &isDown)
	if err != nil {
		return nil, err
	}
	var serversStruct []models.Server
	err = json.Unmarshal(servers, &serversStruct)
	if err != nil {
		return nil, err
	}
	var endpointsStruct []models.Endpoint
	err = json.Unmarshal(endpoints, &endpointsStruct)
	if err != nil {
		return nil, err
	}
	domain := &models.Domain{serversStruct, endpointsStruct, serversChanged, sslGrade, previousSslGrade, logo, title, isDown, id, url, updatedAt}
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
		var servers, endpoints []byte
		var serversChanged, isDown bool
		err := rows.Scan(&id, &servers, &endpoints, &url, &sslGrade, &previousSslGrade, &logo, &title, &updatedAt, &serversChanged, &isDown)
		if err != nil {
			return nil, err
		}
		var serversStruct []models.Server
		decodeErr := json.Unmarshal(servers, &serversStruct)
		if decodeErr != nil {
			return nil, decodeErr
		}
		var endpointsStruct []models.Endpoint
		decodeErr = json.Unmarshal(endpoints, &endpointsStruct)
		if err != nil {
			return nil, err
		}
		domain := &models.Domain{serversStruct, endpointsStruct, serversChanged, sslGrade, previousSslGrade, logo, title, isDown, id, url, updatedAt}
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
	jsonServers, jServerError := json.Marshal(domain.Servers)
	if jServerError != nil {
		return id, jServerError
	}
	jsonEndpoints, jEndpointsError := json.Marshal(domain.Endpoints)
	if jEndpointsError != nil {
		return id, jEndpointsError
	}
	queryErr := r.db.QueryRow(`INSERT INTO domains (servers, endpoints, url, sslGrade, previousSslGrade, logo, title, updatedAt, serversChanged, isDown) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`, jsonServers, jsonEndpoints, domain.Url, domain.SslGrade, domain.PreviousSslGrade, domain.Logo, domain.Title, domain.UpdatedAt, domain.ServersChanged, domain.IsDown).Scan(&id)
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
	jsonServers, jServerError := json.Marshal(domain.Servers)
	if jServerError != nil {
		return id, jServerError
	}
	jsonEndpoints, jEndpointsError := json.Marshal(domain.Endpoints)
	if jEndpointsError != nil {
		return id, jEndpointsError
	}
	_, queryErr := r.db.Exec(`UPDATE domains SET servers=$1, endpoints=$2, url=$3, sslGrade=$4, previousSslGrade=$5, logo=$6, title=$7, updatedAt=$8, serversChanged=$9, isDown=$10 WHERE id=$11`, jsonServers, jsonEndpoints, domain.Url, domain.SslGrade, domain.PreviousSslGrade, domain.Logo, domain.Title, domain.UpdatedAt, domain.ServersChanged, domain.IsDown, domain.Id)
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
	var servers, endpoints []byte
	var serversChanged, isDown bool
	err := r.db.QueryRow("SELECT * FROM domains WHERE url=$1", Url).Scan(&id, &servers, &endpoints, &url, &sslGrade, &previousSslGrade, &logo, &title, &updatedAt, &serversChanged, &isDown)
	if err != nil {
		return nil, err
	}
	var serversStruct []models.Server
	decodeErr := json.Unmarshal(servers, &serversStruct)
	if decodeErr != nil {
		return nil, decodeErr
	}
	var endpointsStruct []models.Endpoint
	decodeErr = json.Unmarshal(endpoints, &endpointsStruct)
	if err != nil {
		return nil, err
	}
	domain := &models.Domain{serversStruct, endpointsStruct, serversChanged, sslGrade, previousSslGrade, logo, title, isDown, id, url, updatedAt}
	return domain, nil
}
