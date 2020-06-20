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

	"github.com/JonatanOrdonez/tr-backend/models"
	"github.com/badoux/goscraper"
	"github.com/likexian/whois-go"
	"github.com/valyala/fasthttp"
)

// DomainService: Structure used to store the domainService functions
type DomainService struct {
}

// NewDomainService: Returns the initialization of the DomainService structure
// Return:
// (*DomainService): Reference to the DomainService object
func NewDomainService() *DomainService {
	return &DomainService{}
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

// RaiseError: Takes two server slices and compares them
// Params:
// (serverA): First server slice
// (serverB): Second server slice
// Return
// (bool): True if the servers are equal. False if not
func (s *DomainService) ServersAreEqual(serversA []models.Server, serversB []models.Server) bool {
	if len(serversA) != len(serversB) {
		return false
	}
	for _, serverA := range serversA {
		serverB, err := findServer(serverA.Address, serversB)
		if err != nil {
			return false
		} else if serverA.SslGrade != serverB.SslGrade {
			return false
		}
	}
	return true
}

// RaiseError: Takes a server address and searches it in a server slice
// Params:
// (address): Server ip address
// (servers): Server slice
// Return
// (models.Server): Server object that was found
// (error): Error if the process fails
func findServer(address string, servers []models.Server) (models.Server, error) {
	for _, server := range servers {
		if server.Address == address {
			return server, nil
		}
	}
	return models.Server{}, errors.New("Server not found")
}
