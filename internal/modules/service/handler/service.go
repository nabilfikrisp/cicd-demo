package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/contract"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

const maxRequestBytes = 16 * 1024

type ServiceHandler struct{ services contract.Usecase }

func NewServiceHandler(services contract.Usecase) *ServiceHandler {
	return &ServiceHandler{services: services}
}

// RegisterRoutes adds the service catalog endpoints to the router.
func (h *ServiceHandler) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/services", h.List)
	routes.POST("/services", h.Create)
}

type serviceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func toServiceResponse(service domain.Service) serviceResponse {
	return serviceResponse{ID: service.ID, Name: service.Name, Description: service.Description}
}

func (h *ServiceHandler) List(c *gin.Context) {
	result, err := h.services.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	response := make([]serviceResponse, 0, len(result))
	for _, service := range result {
		response = append(response, toServiceResponse(service))
	}
	c.JSON(http.StatusOK, response)
}

func (h *ServiceHandler) Create(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if err := json.Unmarshal(body, &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	service, err := h.services.Create(c.Request.Context(), input.Name, input.Description)
	switch {
	case errors.Is(err, domain.ErrInvalidService):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrServiceAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "service already exists"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	default:
		c.JSON(http.StatusCreated, toServiceResponse(service))
	}
}
