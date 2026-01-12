package handlers

import (
	"strings"

	"github.com/captcha-platform/gateway/internal/proxy"
	"github.com/captcha-platform/gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ProxyHandler handles proxying requests to backend services
type ProxyHandler struct {
	authProxy    *proxy.ServiceProxy
	captchaProxy *proxy.ServiceProxy
	logger       *logger.Logger
}

// NewProxyHandler creates a new ProxyHandler
func NewProxyHandler(
	authProxy *proxy.ServiceProxy,
	captchaProxy *proxy.ServiceProxy,
	logger *logger.Logger,
) *ProxyHandler {
	return &ProxyHandler{
		authProxy:    authProxy,
		captchaProxy: captchaProxy,
		logger:       logger,
	}
}

// ProxyToAuth proxies request to auth service
func (h *ProxyHandler) ProxyToAuth(c *gin.Context) {
	// Get the path after /api/v1
	path := c.Request.URL.Path
	
	// Remove /api/v1 prefix to get the service path
	servicePath := strings.TrimPrefix(path, "/api/v1")
	
	h.logger.Debug("Proxying to auth service", "path", servicePath)
	h.authProxy.ProxyRequest(c, servicePath)
}

// ProxyToCaptcha proxies request to captcha service
func (h *ProxyHandler) ProxyToCaptcha(c *gin.Context) {
	// Get the path after /api/v1
	path := c.Request.URL.Path
	
	// Remove /api/v1 prefix to get the service path
	servicePath := strings.TrimPrefix(path, "/api/v1")
	
	h.logger.Debug("Proxying to captcha service", "path", servicePath)
	h.captchaProxy.ProxyRequest(c, servicePath)
}