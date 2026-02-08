package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/models"
	"krizzy/internal/services"
	"krizzy/internal/validation"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ConnectionHandler struct {
	bm *services.BoardManager
}

func NewConnectionHandler(bm *services.BoardManager) *ConnectionHandler {
	return &ConnectionHandler{bm: bm}
}

func (h *ConnectionHandler) ListConnections(c echo.Context) error {
	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	return templates.ConnectionsModal(connections).Render(c.Request().Context(), c.Response().Writer)
}

type CreateConnectionRequest struct {
	Name     string `form:"name"`
	Host     string `form:"host"`
	Port     int    `form:"port"`
	User     string `form:"user"`
	Password string `form:"password"`
	SSLMode  string `form:"ssl_mode"`
}

func (h *ConnectionHandler) CreateConnection(c echo.Context) error {
	var req CreateConnectionRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	req.Name = validation.SanitizeName(req.Name)
	if req.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}
	if req.Host == "" {
		return c.String(http.StatusBadRequest, "Host is required")
	}
	if req.User == "" {
		return c.String(http.StatusBadRequest, "User is required")
	}

	conn := &models.PgConnection{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		SSLMode:  req.SSLMode,
	}

	// Test connectivity before saving
	if err := h.bm.TestConnection(conn); err != nil {
		return c.String(http.StatusBadRequest, "Connection failed: "+err.Error())
	}

	if err := h.bm.PgConnRepo().Create(conn); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save connection")
	}

	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	return templates.ConnectionsList(connections).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ConnectionHandler) TestConnection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid connection ID")
	}

	conn, err := h.bm.PgConnRepo().GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Connection not found")
	}

	if err := h.bm.TestConnection(conn); err != nil {
		return c.String(http.StatusBadRequest, "Connection failed: "+err.Error())
	}

	return c.String(http.StatusOK, "Connection successful")
}

func (h *ConnectionHandler) DeleteConnection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid connection ID")
	}

	// Check if any boards use this connection
	inUse, err := h.bm.HasBoardsUsingConnection(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to check connection usage")
	}
	if inUse {
		return c.String(http.StatusConflict, "Cannot delete: boards are using this connection")
	}

	if err := h.bm.PgConnRepo().Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete connection")
	}

	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	return templates.ConnectionsList(connections).Render(c.Request().Context(), c.Response().Writer)
}
