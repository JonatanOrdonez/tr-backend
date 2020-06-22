package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	interfaces "github.com/JonatanOrdonez/tr-backend/interfaces"

	"github.com/JonatanOrdonez/tr-backend/models"
	"github.com/badoux/goscraper"
	"github.com/likexian/whois-go"
	"github.com/valyala/fasthttp"
)

// DomainService: Structure used to store the domainService functions
type DomainService struct {
	domainRepo interfaces.IDomainRepository
}

// NewDomainService: Receives a reference to the domainRepo interface and stores it in the DomainService structure
// Params:
// (domainRepo): Reference to a domainRepo interface
// Return:
// (*DomainService): Reference to the BaseHandler object
func NewDomainService(domainRepo interfaces.IDomainRepository) *DomainService {
	return &DomainService{domainRepo: domainRepo}
}

// ResponseDomains: Returns a JSON object domain Slice
// Return:
// ([]byte): JSON object
// (error): Error if the process fails
func (s *DomainService) GetDomains() ([]byte, error) {
	domains, err := s.domainRepo.GetAll()
	if err != nil {
		return nil, err
	}
	items := &models.Items{domains}
	jsonBody, jsonError := json.Marshal(items)
	if jsonError != nil {
		return nil, jsonError
	}
	return jsonBody, nil
}

// CheckDomain: Checks if the domain exists
// If the request has the host query param empty (""), the redirection is made to ResponseDomains
// If a domain is not found that matches its url as host, then the redirection is made to AddDomain function
// If there is a domain such that the url equals host, then the redirection is made to UpdateDomain function
// Params:
// (hostPath): Host value of the path param
// Return:
// ([]byte): JSON object
// (error): Error if the process fails
func (s *DomainService) CheckDomain(hostPath string) ([]byte, error) {
	domain, domainErr := s.domainRepo.FindByUrl(hostPath)
	if domainErr != nil {
		return s.AddDomain(hostPath)
	} else {
		return s.UpdateDomain(hostPath, domain)
	}
}

// AddDomain: Creates a new domain in the database, according to the requirements of the test.
// Params:
// (hostPath): Host value of the path param
// Return:
// ([]byte): JSON object
// (error): Error if the process fails
func (s *DomainService) AddDomain(hostPath string) ([]byte, error) {
	var ssllabs *models.Ssllabs
	var ssllabsError error
	ssllabs, ssllabsError = s.CheckDomainInSsllabs(hostPath)
	if ssllabsError != nil {
		return nil, ssllabsError
	}
	if ssllabs.Status == "ERROR" {
		servers := []models.Server{}
		UpdatedAt := time.Now().Unix()
		newDomain := &models.Domain{servers, ssllabs.Endpoints, false, "", "", "", "", true, 1, hostPath, UpdatedAt}
		id, saveDomainError := s.domainRepo.Save(newDomain)
		if saveDomainError != nil {
			return nil, saveDomainError
		}
		domainEntity, queryErr := s.domainRepo.FindByID(id)
		if queryErr != nil {
			return nil, queryErr
		}
		jsonBody, jsonError := json.Marshal(domainEntity)
		if jsonError != nil {
			return nil, jsonError
		}
		return jsonBody, nil
	}
	if ssllabs.Status == "DNS" {
		ssllabs, ssllabsError = s.CheckDomainInSsllabs(hostPath)
	}
	if ssllabsError != nil {
		return nil, errors.New(ssllabs.StatusMessage)
	}
	if ssllabs.Status == "ERROR" {
		servers := []models.Server{}
		endpoints := []models.Endpoint{}
		UpdatedAt := time.Now().Unix()
		newDomain := &models.Domain{servers, endpoints, false, "", "", "", "", true, 1, hostPath, UpdatedAt}
		id, saveDomainError := s.domainRepo.Save(newDomain)
		if saveDomainError != nil {
			return nil, saveDomainError
		}
		domainEntity, queryErr := s.domainRepo.FindByID(id)
		if queryErr != nil {
			return nil, queryErr
		}
		jsonBody, jsonError := json.Marshal(domainEntity)
		if jsonError != nil {
			return nil, jsonError
		}
		return jsonBody, nil
	}
	servers, fetchSDError := s.FetchServersData(ssllabs.Endpoints)
	if fetchSDError != nil {
		return nil, fetchSDError
	}
	sslGrade := ""
	lowerServer, lsErr := s.GetLowerServer(servers)
	if lsErr == nil {
		sslGrade = lowerServer.SslGrade
	}
	logo, title, _ := s.ScrapPage(hostPath)
	UpdatedAt := time.Now().Unix()
	newDomain := &models.Domain{servers, ssllabs.Endpoints, false, sslGrade, sslGrade, logo, title, false, 1, hostPath, UpdatedAt}
	id, saveDomainError := s.domainRepo.Save(newDomain)
	if saveDomainError != nil {
		return nil, saveDomainError
	}
	domainEntity, queryErr := s.domainRepo.FindByID(id)
	if queryErr != nil {
		return nil, queryErr
	}
	jsonBody, jsonError := json.Marshal(domainEntity)
	if jsonError != nil {
		return nil, jsonError
	}
	return jsonBody, nil
}

// UpdateDomain: Update a new domain in the database, according to the requirements of the test.
// Params:
// (hostPath): Host value of the path param
// (domain): Reference to the domain
// ([]byte): JSON object
// (error): Error if the process fails
func (s *DomainService) UpdateDomain(hostPath string, domain *models.Domain) ([]byte, error) {
	var ssllabs *models.Ssllabs
	var ssllabsError error
	ssllabs, ssllabsError = s.CheckDomainInSsllabs(hostPath)
	if ssllabsError != nil {
		return nil, ssllabsError
	}
	if ssllabs.Status == "ERROR" {
		jsonBody, jsonError := json.Marshal(domain)
		if jsonError != nil {
			return nil, jsonError
		}
		return jsonBody, nil
	}
	if ssllabs.Status == "DNS" {
		ssllabs, ssllabsError = s.CheckDomainInSsllabs(hostPath)
	}
	if ssllabsError != nil {
		return nil, ssllabsError
	}
	if ssllabs.Status == "ERROR" {
		jsonBody, jsonError := json.Marshal(domain)
		if jsonError != nil {
			return nil, jsonError
		}
		return jsonBody, nil
	}
	endpointsAreEquals := s.EndpointsAreEqual(domain.Endpoints, ssllabs.Endpoints)
	currentDate := time.Unix(time.Now().Unix(), 0)
	updatedAt := time.Unix(domain.UpdatedAt, 0)
	elapsed := currentDate.Sub(updatedAt).Hours()
	if endpointsAreEquals == false && elapsed < 1 {
		servers, fetchSDError := s.FetchServersData(ssllabs.Endpoints)
		if fetchSDError != nil {
			return nil, fetchSDError
		}
		lowerServer, lsErr := s.GetLowerServer(servers)
		sslGrade := ""
		if lsErr == nil {
			sslGrade = lowerServer.SslGrade
		}
		newUpdatedAt := time.Now().Unix()
		newDomain := &models.Domain{servers, ssllabs.Endpoints, true, sslGrade, domain.SslGrade, domain.Logo, domain.Title, false, domain.Id, hostPath, newUpdatedAt}
		id, updatedErr := s.domainRepo.Update(newDomain)
		if updatedErr != nil {
			return nil, updatedErr
		}
		domainEntity, queryErr := s.domainRepo.FindByID(id)
		if queryErr != nil {
			return nil, queryErr
		}
		jsonBody, jsonError := json.Marshal(domainEntity)
		if jsonError != nil {
			return nil, jsonError
		}
		return jsonBody, nil
	} else {
		domain.ServersChanged = false
		id, updatedErr := s.domainRepo.Update(domain)
		if updatedErr != nil {
			return nil, updatedErr
		}
		domainEntity, queryErr := s.domainRepo.FindByID(id)
		if queryErr != nil {
			return nil, queryErr
		}
		jsonBody, jsonError := json.Marshal(domainEntity)
		if jsonError != nil {
			return nil, jsonError
		}
		return jsonBody, nil
	}
}

// CheckDomainInSsllabs: Takes a domain url and makes a request to the Ssllabs api for obtain information
// Params:
// (url): URl of the domain you are looking for
// Return:
// (*models.Ssllabs): Reference to the response object
// (error): Error if the process fails
func (s *DomainService) CheckDomainInSsllabs(url string) (*models.Ssllabs, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.ssllabs.com/api/v3/analyze?host=%s", url))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		var ssllabs *models.Ssllabs
		unmarshalError := json.Unmarshal(body, &ssllabs)
		if unmarshalError != nil {
			return nil, unmarshalError
		}
		return ssllabs, nil
	}
	return nil, errors.New("Invalid host")
}

// GetLowerServer: Takes a server slice and looks for the one with the lowest SSL grade
// Params:
// ([]models.Server): Server slice
// Return:
// (*models.Server): Reference to the lowest server
// (error): Error if the process fails
func (s *DomainService) GetLowerServer(servers []models.Server) (*models.Server, error) {
	if len(servers) == 0 {
		return nil, errors.New("servers array is empty")
	}
	var lowerServer *models.Server
	for _, server := range servers {
		if lowerServer == nil {
			lowerServer = &server
		}
		lowerGrade := 0
		currentGrade := 0
		if lowerServer.SslGrade != "" {
			lowerGrade = int([]rune(lowerServer.SslGrade)[0])
		}
		if server.SslGrade != "" {
			currentGrade = int([]rune(server.SslGrade)[0])
		}
		if (lowerGrade - currentGrade) > 0 {
			lowerServer = &server
		}
	}
	return lowerServer, nil
}

// FetchServersData: Takes an endpoint slice and converts it to a server slice
// Params:
// ([]models.Endpoint): Endpoint slice
// Return:
// ([]models.Server): Server slice
// (error): Error if the process fails
func (s *DomainService) FetchServersData(endpoints []models.Endpoint) ([]models.Server, error) {
	servers := make([]models.Server, 0)
	parallelization := 3
	c := make(chan models.Endpoint)
	var wg sync.WaitGroup
	wg.Add(parallelization)
	for ii := 0; ii < parallelization; ii++ {
		go func(c chan models.Endpoint) {
			for {
				v, more := <-c
				if more == false {
					wg.Done()
					return
				}
				server, err := fetchServerData(v)
				if err == nil {
					servers = append(servers, *server)
				}
			}
		}(c)
	}
	for _, endpoint := range endpoints {
		c <- endpoint
	}
	close(c)
	wg.Wait()
	return servers, nil
}

// fetchServerData: Auxiliary function that takes an endpoint object and converts it to a server object
// Params:
// (models.Endpoint): Endpoint object
// Return:
// (models.Server): Server reference
// (error): Error if the process fails
func fetchServerData(endpoint models.Endpoint) (*models.Server, error) {
	whoisRawData, errRawData := whois.Whois(endpoint.IpAddress)
	if errRawData != nil {
		return nil, errRawData
	}
	countryChanel := make(chan string)
	orgChanel := make(chan string)
	go findWordInData("Country", whoisRawData, countryChanel)
	go findWordInData("OrgName", whoisRawData, orgChanel)
	country, owner := <-countryChanel, <-orgChanel
	server := &models.Server{endpoint.IpAddress, endpoint.Grade, country, owner}
	return server, nil
}

// fetchServerData: Auxiliary function that takes a word and searches it in a text through the regex function
// Params:
// (word): Word to be looked up
// (rawData): Search text
// (chanel): Channel through which the match is received
func findWordInData(word string, rawData string, chanel chan string) {
	regexClient := regexp.MustCompile(fmt.Sprintf(`%s:[\w .]*`, word))
	strSplit := strings.Split(regexClient.FindString(rawData), ":")
	if len(strSplit) == 2 {
		chanel <- strings.Trim(strSplit[1], " ")
	}
	chanel <- ""
}

// ScrapPage: Auxiliary function that takes a domain and returns the icon and title of the page
// Params:
// (url): URl of the domain you are looking for
// Return:
// (logo): Web page logo
// (title): Web page title
// (error): Error if the process fails
func (s *DomainService) ScrapPage(url string) (logo string, title string, err error) {
	domain := url
	if strings.HasPrefix(domain, "ht") == false {
		domain = fmt.Sprintf("http://%s", domain)
	}
	scraper, err := goscraper.Scrape(domain, 5)
	if err != nil {
		return "", "", nil
	}
	return scraper.Preview.Icon, scraper.Preview.Title, nil
}

// RaiseError: Takes a ctx reference and responses a JSON error to the client
// Params:
// (ctx): Request reference
// (errorCode): Error code
// (errorMessage): Error message
func (s *DomainService) RaiseError(ctx *fasthttp.RequestCtx, errorCode int, errorMessage string) {
	errorEntity := &models.Error{errorCode, errorMessage}
	jsonBody, _ := json.Marshal(errorEntity)
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(errorCode)
	ctx.Response.SetBody(jsonBody)
}

// RaiseError: Takes two endpoint slices and compares them
// Params:
// (endpointA): First endpoint slice
// (endpointB): Second endpoint slice
// Return
// (bool): True if the endpoint slices are equal. False if not
func (s *DomainService) EndpointsAreEqual(endpointsA []models.Endpoint, endpointsB []models.Endpoint) bool {
	if len(endpointsA) != len(endpointsB) {
		return false
	}
	for _, endpointA := range endpointsA {
		endpointB, err := findEndpoint(endpointA.IpAddress, endpointsB)
		if err != nil {
			return false
		}
		if compareEndpoints(endpointA, endpointB) == false {
			return false
		}
	}
	return true
}

// findEndpoint: Takes a endpoint ip address and searches it into a endpoint slice
// Params:
// (IpAddress): Endpoint ip address
// (endpoints): Endpoint slice
// Return
// (models.Endpoint): Endpoint object that was found
// (error): Error if the process fails
func findEndpoint(IpAddress string, endpoints []models.Endpoint) (models.Endpoint, error) {
	for _, endpoint := range endpoints {
		if IpAddress == endpoint.IpAddress {
			return endpoint, nil
		}
	}
	return models.Endpoint{}, errors.New("Endpoint not found")
}

// compareEndpoints: Takes two endpoint objects and compares them
// Params:
// (endpointA): First Endpoint object
// (endpointB): Second Endpoint object
// Return
// (bool): True if endpoints are equals. False if not
func compareEndpoints(endpointA models.Endpoint, endpointB models.Endpoint) bool {
	if endpointA.ServerName != endpointB.ServerName {
		return false
	}
	if endpointA.StatusMessage != endpointB.StatusMessage {
		return false
	}
	if endpointA.Grade != endpointB.Grade {
		return false
	}
	if endpointA.GradeTrustIgnored != endpointB.GradeTrustIgnored {
		return false
	}
	if endpointA.IsExceptional != endpointB.IsExceptional {
		return false
	}
	if endpointA.Progress != endpointB.Progress {
		return false
	}
	if endpointA.Duration != endpointB.Duration {
		return false
	}
	if endpointA.StatusDetails != endpointB.StatusDetails {
		return false
	}
	if endpointA.StatusDetailsMessage != endpointB.StatusDetailsMessage {
		return false
	}
	if endpointA.Delegation != endpointB.Delegation {
		return false
	}
	return true
}
