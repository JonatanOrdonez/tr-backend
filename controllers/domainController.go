package controllers

import (
	"encoding/json"
	"time"

	interfaces "github.com/JonatanOrdonez/tr-backend/interfaces"
	"github.com/JonatanOrdonez/tr-backend/models"
	"github.com/valyala/fasthttp"
)

// BaseHandler: Structure used to store a domainRepo and domainService object
type BaseHandler struct {
	domainRepo    interfaces.IDomainRepository
	domainService interfaces.IDomainService
}

// NewDomainHandler: Receives a reference to the domainRepo and domainService interfaces and stores them in the BaseHandler structure
// Params:
// (domainRepo): Reference to a domainRepo interface
// (domainService): Reference to a domainService interface
// Return:
// (*BaseHandler): Reference to the BaseHandler object
func NewDomainHandler(domainRepo interfaces.IDomainRepository, domainService interfaces.IDomainService) *BaseHandler {
	return &BaseHandler{domainRepo: domainRepo, domainService: domainService}
}

// GetDomains: Returns a JSON http response with the domain Slice
// Params:
// (ctx): Request reference
func (h *BaseHandler) GetDomains(ctx *fasthttp.RequestCtx) {
	domains, err := h.domainRepo.GetAll()
	if err != nil {
		h.domainService.RaiseError(ctx, 400, err.Error())
	}
	jsonBody, jsonError := json.Marshal(domains)
	if jsonError != nil {
		h.domainService.RaiseError(ctx, 400, jsonError.Error())
	}
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(200)
	ctx.Response.SetBody(jsonBody)
}

// CheckDomain: Handles the request that gets at the endpoint /api/v1/analyze.
// If the request has the host query param empty (""), the redirection is made to GetDomains
// If a domain is not found that matches its url as host, then the redirection is made to addDomain
// If there is a domain such that the url equals host, then the redirection is made to updateDomain
// Params:
// (ctx): Request reference
func (h *BaseHandler) CheckDomain(ctx *fasthttp.RequestCtx) {
	hostPath := string(ctx.QueryArgs().Peek("host"))
	if hostPath == "" {
		h.GetDomains(ctx)
		return
	}
	domain, domainErr := h.domainRepo.FindByUrl(hostPath)
	if domainErr != nil {
		h.addDomain(ctx, hostPath)
		return
	}
	h.updateDomain(ctx, hostPath, domain)
	return
}

// addDomain: Creates a new domain in the database, according to the requirements of the test.
// Params:
// (ctx): Request reference
// (hostPath): Host value of the path param
func (h *BaseHandler) addDomain(ctx *fasthttp.RequestCtx, hostPath string) {
	var ssllabs *models.Ssllabs
	var ssllabsError error
	ssllabs, ssllabsError = h.domainService.CheckDomainInSsllabs(hostPath)
	if ssllabsError != nil {
		h.domainService.RaiseError(ctx, 400, ssllabsError.Error())
		return
	}
	if ssllabs.Status == "ERROR" {
		h.domainService.RaiseError(ctx, 400, ssllabs.StatusMessage)
		return
	}
	if ssllabs.Status == "DNS" {
		ssllabs, ssllabsError = h.domainService.CheckDomainInSsllabs(hostPath)
	}
	if ssllabsError != nil {
		h.domainService.RaiseError(ctx, 400, ssllabsError.Error())
		return
	}
	if ssllabs.Status == "ERROR" {
		h.domainService.RaiseError(ctx, 400, ssllabs.StatusMessage)
		return
	}
	servers, fetchSDError := h.domainService.FetchServersData(ssllabs.Endpoints)
	if fetchSDError != nil {
		h.domainService.RaiseError(ctx, 200, fetchSDError.Error())
		return
	}
	sslGrade := ""
	lowerServer, lsErr := h.domainService.GetLowerServer(servers)
	if lsErr == nil {
		sslGrade = lowerServer.SslGrade
	}
	isDown := true
	if ssllabs.Status == "READY" {
		isDown = false
	}
	logo, title, _ := h.domainService.ScrapPage(hostPath)
	UpdatedAt := time.Now().Unix()
	newDomain := &models.Domain{servers, false, sslGrade, sslGrade, logo, title, isDown, 1, hostPath, UpdatedAt}
	id, saveDomainError := h.domainRepo.Save(newDomain)
	if saveDomainError != nil {
		h.domainService.RaiseError(ctx, 400, saveDomainError.Error())
		return
	}
	domainEntity, queryErr := h.domainRepo.FindByID(id)
	if queryErr != nil {
		h.domainService.RaiseError(ctx, 400, queryErr.Error())
		return
	}
	jsonBody, jsonError := json.Marshal(domainEntity)
	if jsonError != nil {
		h.domainService.RaiseError(ctx, 400, jsonError.Error())
		return
	}
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(200)
	ctx.Response.SetBody(jsonBody)
}

// addDomain: Update a new domain in the database, according to the requirements of the test.
// Params:
// (ctx): Request reference
// (hostPath): Host value of the path param
// (domain): Reference to the domain
func (h *BaseHandler) updateDomain(ctx *fasthttp.RequestCtx, hostPath string, domain *models.Domain) {
	var ssllabs *models.Ssllabs
	var ssllabsError error
	ssllabs, ssllabsError = h.domainService.CheckDomainInSsllabs(hostPath)
	if ssllabsError != nil {
		h.domainService.RaiseError(ctx, 400, ssllabsError.Error())
		return
	}
	if ssllabs.Status == "ERROR" {
		h.domainService.RaiseError(ctx, 400, ssllabs.StatusMessage)
		return
	}
	if ssllabs.Status == "DNS" {
		ssllabs, ssllabsError = h.domainService.CheckDomainInSsllabs(hostPath)
		return
	}
	if ssllabsError != nil {
		h.domainService.RaiseError(ctx, 400, ssllabsError.Error())
		return
	}
	if ssllabs.Status == "ERROR" {
		h.domainService.RaiseError(ctx, 400, ssllabs.StatusMessage)
		return
	}
	currentServers := domain.Servers
	servers, fetchSDError := h.domainService.FetchServersData(ssllabs.Endpoints)
	if fetchSDError != nil {
		h.domainService.RaiseError(ctx, 400, fetchSDError.Error())
		return
	}
	serversAreEqual := h.domainService.ServersAreEqual(currentServers, servers)
	currentDate := time.Unix(time.Now().Unix(), 0)
	updatedAt := time.Unix(domain.UpdatedAt, 0)
	elapsed := currentDate.Sub(updatedAt).Hours()
	serversChanged := false
	if serversAreEqual && elapsed < 0 {
		sslGrade := ""
		lowerServer, lsErr := h.domainService.GetLowerServer(servers)
		if lsErr == nil {
			sslGrade = lowerServer.SslGrade
		}
		isDown := true
		if ssllabs.Status == "READY" {
			isDown = false
		}
		newUpdatedAt := time.Now().Unix()
		newDomain := &models.Domain{servers, serversChanged, sslGrade, domain.SslGrade, domain.Logo, domain.Title, isDown, domain.Id, hostPath, newUpdatedAt}
		id, updatedErr := h.domainRepo.Update(newDomain)
		if updatedErr != nil {
			h.domainService.RaiseError(ctx, 400, updatedErr.Error())
			return
		}
		domainEntity, queryErr := h.domainRepo.FindByID(id)
		if queryErr != nil {
			h.domainService.RaiseError(ctx, 400, queryErr.Error())
			return
		}
		jsonBody, jsonError := json.Marshal(domainEntity)
		if jsonError != nil {
			h.domainService.RaiseError(ctx, 400, jsonError.Error())
			return
		}
		ctx.SetContentType("application/json; charset=utf-8")
		ctx.SetStatusCode(200)
		ctx.Response.SetBody(jsonBody)
	} else {
		jsonBody, jsonError := json.Marshal(domain)
		if jsonError != nil {
			h.domainService.RaiseError(ctx, 400, jsonError.Error())
			return
		}
		ctx.SetContentType("application/json; charset=utf-8")
		ctx.SetStatusCode(200)
		ctx.Response.SetBody(jsonBody)
	}
}
