package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/service"
	"meos-graphics/internal/state"
)

type Handler struct {
	service *service.Service
}

func New(appState *state.State) *Handler {
	return &Handler{
		service: service.New(appState),
	}
}

// GetClasses returns all competition classes
// @Summary Get all competition classes
// @Description Get a list of all competition classes sorted by order key
// @Tags classes
// @Accept json
// @Produce json
// @Success 200 {array} service.ClassInfo
// @Router /classes [get]
func (h *Handler) GetClasses(c *gin.Context) {
	classes := h.service.GetClasses()
	c.JSON(http.StatusOK, classes)
}

// GetStartList returns the start list for a specific class
// @Summary Get start list for a class
// @Description Get the start list for a specific competition class
// @Tags classes
// @Accept json
// @Produce json
// @Param classId path int true "Class ID"
// @Success 200 {array} service.StartListEntry
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /classes/{classId}/startlist [get]
func (h *Handler) GetStartList(c *gin.Context) {
	var classID int
	if _, err := fmt.Sscanf(c.Param("classId"), "%d", &classID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	startList, err := h.service.GetStartList(classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, startList)
}

// GetResults returns the results for a specific class
// @Summary Get results for a class
// @Description Get the results for a specific competition class including positions and times
// @Tags classes
// @Accept json
// @Produce json
// @Param classId path int true "Class ID"
// @Success 200 {array} service.ResultEntry
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /classes/{classId}/results [get]
func (h *Handler) GetResults(c *gin.Context) {
	var classID int
	if _, err := fmt.Sscanf(c.Param("classId"), "%d", &classID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	results, err := h.service.GetResults(classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetSplits returns split times for a specific class
// @Summary Get split times for a class
// @Description Get split times at each control for a specific competition class
// @Tags classes
// @Accept json
// @Produce json
// @Param classId path int true "Class ID"
// @Success 200 {object} service.SplitsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /classes/{classId}/splits [get]
func (h *Handler) GetSplits(c *gin.Context) {
	var classID int
	if _, err := fmt.Sscanf(c.Param("classId"), "%d", &classID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	splits, err := h.service.GetSplits(classID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, splits)
}
