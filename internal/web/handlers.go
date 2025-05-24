package web

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/service"
)

// Handler handles web page requests
type Handler struct {
	service   *service.Service
	templates *template.Template
}

// New creates a new web handler
func New(svc *service.Service) *Handler {
	h := &Handler{
		service: svc,
	}
	h.loadTemplates()
	return h
}

// loadTemplates loads HTML templates
func (h *Handler) loadTemplates() {
	h.templates = GetTemplates()
}

// HomePage serves the main web interface
func (h *Handler) HomePage(c *gin.Context) {
	classes := h.service.GetClasses()

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(c.Writer, "index", gin.H{
		"Title":   "MeOS Graphics",
		"Classes": classes,
	}); err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
	}
}

// ClassPage serves the page for a specific class
func (h *Handler) ClassPage(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Error": "Invalid class ID",
		})
		return
	}

	// Get class info
	classes := h.service.GetClasses()
	var className string
	for _, class := range classes {
		if class.ID == classID {
			className = class.Name
			break
		}
	}

	if className == "" {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Error": "Class not found",
		})
		return
	}

	c.HTML(http.StatusOK, "class", gin.H{
		"Title":     className,
		"ClassID":   classID,
		"ClassName": className,
	})
}

// StartListPartial serves the start list as an HTML partial for HTMX
func (h *Handler) StartListPartial(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error-partial", gin.H{
			"Error": "Invalid class ID",
		})
		return
	}

	startList, err := h.service.GetStartList(classID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-partial", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "startlist-partial", gin.H{
		"StartList": startList,
	})
}

// ResultsPartial serves the results as an HTML partial for HTMX
func (h *Handler) ResultsPartial(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error-partial", gin.H{
			"Error": "Invalid class ID",
		})
		return
	}

	results, err := h.service.GetResults(classID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-partial", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "results-partial", gin.H{
		"Results": results,
	})
}

// SplitsPartial serves the splits as an HTML partial for HTMX
func (h *Handler) SplitsPartial(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error-partial", gin.H{
			"Error": "Invalid class ID",
		})
		return
	}

	splits, err := h.service.GetSplits(classID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-partial", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "splits-partial", gin.H{
		"Splits": splits,
	})
}
