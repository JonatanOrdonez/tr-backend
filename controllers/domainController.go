package controllers

import (
	interfaces "github.com/JonatanOrdonez/tr-backend/interfaces"
	"github.com/valyala/fasthttp"
)

// BaseHandler: Structure used to store a domainRepo and domainService object
type BaseHandler struct {
	domainService interfaces.IDomainService
}

// NewDomainHandler: Receives a reference to the domainRepo and domainService interfaces and stores them in the BaseHandler structure
// Params:
// (domainRepo): Reference to a domainRepo interface
// (domainService): Reference to a domainService interface
// Return:
// (*BaseHandler): Reference to the BaseHandler object
func NewDomainController(domainService interfaces.IDomainService) *BaseHandler {
	return &BaseHandler{domainService: domainService}
}

// ResponseCheckDomain: Handles the request that gets at the endpoint /api/v1/analyze.
// If the request has the host query param empty (""), the redirection is made to ResponseDomains
// Else, call the function CheckDomain from the DomainService interface
// Params:
// (ctx): Request reference
func (h *BaseHandler) ResponseCheckDomain(ctx *fasthttp.RequestCtx) {
	hostPath := string(ctx.QueryArgs().Peek("host"))
	if hostPath == "" {
		h.ResponseDomains(ctx)
	} else {
		jsonBody, domainErr := h.domainService.CheckDomain(hostPath)
		if domainErr != nil {
			h.domainService.RaiseError(ctx, 400, domainErr.Error())
		} else {
			ctx.SetContentType("application/json; charset=utf-8")
			ctx.SetStatusCode(200)
			ctx.Response.SetBody(jsonBody)
		}
	}
}

// ResponseDomains: Returns a JSON http response with the domain Slice
// Params:
// (ctx): Request reference
func (h *BaseHandler) ResponseDomains(ctx *fasthttp.RequestCtx) {
	jsonDomains, err := h.domainService.GetDomains()
	if err != nil {
		h.domainService.RaiseError(ctx, 400, err.Error())
	} else {
		ctx.SetContentType("application/json; charset=utf-8")
		ctx.SetStatusCode(200)
		ctx.Response.SetBody(jsonDomains)
	}
}
