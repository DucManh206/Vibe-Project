package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/captcha-platform/gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ServiceProxy handles proxying requests to backend services
type ServiceProxy struct {
	targetURL  string
	httpClient *http.Client
	logger     *logger.Logger
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(targetURL string, logger *logger.Logger) *ServiceProxy {
	return &ServiceProxy{
		targetURL: targetURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// ProxyRequest proxies the request to the target service
func (p *ServiceProxy) ProxyRequest(c *gin.Context, path string) {
	// Build target URL
	targetURL, err := url.Parse(p.targetURL)
	if err != nil {
		p.logger.Error("Failed to parse target URL", "error", err, "url", p.targetURL)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "proxy_error",
			"message": "Failed to proxy request",
		})
		return
	}

	// Set the path
	targetURL.Path = "/api/v1" + path
	targetURL.RawQuery = c.Request.URL.RawQuery

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			p.logger.Error("Failed to read request body", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Failed to read request body",
			})
			return
		}
	}

	// Create proxy request
	proxyReq, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		targetURL.String(),
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		p.logger.Error("Failed to create proxy request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "proxy_error",
			"message": "Failed to create proxy request",
		})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Forward client IP
	clientIP := c.ClientIP()
	if clientIP != "" {
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
		proxyReq.Header.Set("X-Real-IP", clientIP)
	}

	// Forward request ID
	if requestID := c.GetString("request_id"); requestID != "" {
		proxyReq.Header.Set("X-Request-ID", requestID)
	}

	// Forward user info if authenticated
	if userID, exists := c.Get("user_id"); exists {
		proxyReq.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
	}
	if userEmail, exists := c.Get("user_email"); exists {
		proxyReq.Header.Set("X-User-Email", userEmail.(string))
	}
	if userRole, exists := c.Get("user_role"); exists {
		proxyReq.Header.Set("X-User-Role", userRole.(string))
	}

	// Forward API key hash if present
	if apiKeyHash, exists := c.Get("api_key_hash"); exists {
		proxyReq.Header.Set("X-API-Key-Hash", apiKeyHash.(string))
	}

	// Execute request
	p.logger.Debug("Proxying request",
		"method", c.Request.Method,
		"path", path,
		"target", targetURL.String(),
	)

	resp, err := p.httpClient.Do(proxyReq)
	if err != nil {
		p.logger.Error("Failed to proxy request", "error", err, "target", targetURL.String())
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "service_unavailable",
			"message": "Backend service is unavailable",
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read response body", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "proxy_error",
			"message": "Failed to read response",
		})
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Write response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// HealthCheck checks if the target service is healthy
func (p *ServiceProxy) HealthCheck() error {
	targetURL, err := url.Parse(p.targetURL)
	if err != nil {
		return err
	}
	targetURL.Path = "/health"

	resp, err := p.httpClient.Get(targetURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}