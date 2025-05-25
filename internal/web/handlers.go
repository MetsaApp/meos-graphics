package web

import (
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"

	"meos-graphics/internal/service"
	"meos-graphics/internal/web/templates"
)

// Handler handles web page requests
type Handler struct {
	service           *service.Service
	simulationEnabled bool
}

// New creates a new web handler
func New(svc *service.Service, simulationEnabled bool) *Handler {
	return &Handler{
		service:           svc,
		simulationEnabled: simulationEnabled,
	}
}

// renderTempl is a helper function to render templ components
func renderTempl(c *gin.Context, status int, component templ.Component) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Template render error: %v", err)
	}
}

// HomePage serves the main web interface
func (h *Handler) HomePage(c *gin.Context) {
	classes := h.service.GetClasses()
	renderTempl(c, http.StatusOK, templates.HomePage(classes, h.simulationEnabled))
}

// ClassPage serves the page for a specific class
func (h *Handler) ClassPage(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		renderTempl(c, http.StatusBadRequest, templates.ErrorPage("Invalid class ID"))
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
		renderTempl(c, http.StatusNotFound, templates.ErrorPage("Class not found"))
		return
	}

	renderTempl(c, http.StatusOK, templates.ClassPage(classID, className, h.simulationEnabled))
}

// StartListPartial serves the start list as an HTML partial for HTMX
func (h *Handler) StartListPartial(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		renderTempl(c, http.StatusBadRequest, templates.ErrorPartial("Invalid class ID"))
		return
	}

	startList, err := h.service.GetStartList(classID)
	if err != nil {
		renderTempl(c, http.StatusInternalServerError, templates.ErrorPartial(err.Error()))
		return
	}

	renderTempl(c, http.StatusOK, templates.StartListPartial(startList))
}

// ResultsPartial serves the results as an HTML partial for HTMX
func (h *Handler) ResultsPartial(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		renderTempl(c, http.StatusBadRequest, templates.ErrorPartial("Invalid class ID"))
		return
	}

	results, err := h.service.GetResults(classID)
	if err != nil {
		renderTempl(c, http.StatusInternalServerError, templates.ErrorPartial(err.Error()))
		return
	}

	renderTempl(c, http.StatusOK, templates.ResultsPartial(results))
}

// SplitsPartial serves the splits as an HTML partial for HTMX
func (h *Handler) SplitsPartial(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("classId"))
	if err != nil {
		renderTempl(c, http.StatusBadRequest, templates.ErrorPartial("Invalid class ID"))
		return
	}

	splits, err := h.service.GetSplits(classID)
	if err != nil {
		renderTempl(c, http.StatusInternalServerError, templates.ErrorPartial(err.Error()))
		return
	}

	renderTempl(c, http.StatusOK, templates.SplitsPartial(*splits))
}
