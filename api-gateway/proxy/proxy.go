package proxy

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ServiceProxy handles proxying requests to backend services
type ServiceProxy struct {
	authServiceURL   string
	chatServiceURL   string
	socialServiceURL string
	aiServiceURL     string
	client           *http.Client
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(authURL, chatURL, socialURL, aiURL string) *ServiceProxy {
	return &ServiceProxy{
		authServiceURL:   authURL,
		chatServiceURL:   chatURL,
		socialServiceURL: socialURL,
		aiServiceURL:     aiURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest proxies a request to the target service
func (p *ServiceProxy) ProxyRequest(c *fiber.Ctx, targetURL string) error {
	// Build target URL
	path := c.Path()
	fullURL := targetURL + path

	// Add query parameters
	if len(c.Request().URI().QueryString()) > 0 {
		fullURL += "?" + string(c.Request().URI().QueryString())
	}

	// Create new request
	req, err := http.NewRequestWithContext(
		c.Context(),
		c.Method(),
		fullURL,
		strings.NewReader(string(c.Body())),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create proxy request",
		})
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "Service temporarily unavailable",
		})
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to read service response",
		})
	}

	return c.Status(resp.StatusCode).Send(body)
}

// ProxyToAuth proxies to auth service
func (p *ServiceProxy) ProxyToAuth(c *fiber.Ctx) error {
	return p.ProxyRequest(c, p.authServiceURL)
}

// ProxyToChat proxies to chat service
func (p *ServiceProxy) ProxyToChat(c *fiber.Ctx) error {
	return p.ProxyRequest(c, p.chatServiceURL)
}

// ProxyToSocial proxies to social service
func (p *ServiceProxy) ProxyToSocial(c *fiber.Ctx) error {
	return p.ProxyRequest(c, p.socialServiceURL)
}

// ProxyToAI proxies to AI service
func (p *ServiceProxy) ProxyToAI(c *fiber.Ctx) error {
	return p.ProxyRequest(c, p.aiServiceURL)
}
