package interfaces

import (
	"github.com/JonatanOrdonez/tr-backend/models"
	"github.com/valyala/fasthttp"
)

// IDomainService...
type IDomainService interface {
	CheckDomainInSsllabs(url string) (*models.Ssllabs, error)
	FetchServersData(endpoints []models.Endpoint) ([]models.Server, error)
	ScrapPage(url string) (logo string, title string, err error)
	RaiseError(ctx *fasthttp.RequestCtx, errorCode int, errorMessage string)
	ServersAreEqual(serversA []models.Server, serversB []models.Server) bool
	GetLowerServer(servers []models.Server) (*models.Server, error)
	GetDomains() ([]byte, error)
	CheckDomain(hostPath string) ([]byte, error)
}
