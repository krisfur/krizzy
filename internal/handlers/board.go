package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/services"
	"krizzy/internal/validation"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type BoardHandler struct {
	bm *services.BoardManager
}

func NewBoardHandler(bm *services.BoardManager) *BoardHandler {
	return &BoardHandler{bm: bm}
}

// ListBoards shows all boards
func (h *BoardHandler) ListBoards(c echo.Context) error {
	boards, err := h.bm.GetAllBoards()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load boards")
	}

	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return templates.BoardsList(boards, connections).Render(c.Request().Context(), c.Response().Writer)
	}

	return templates.BoardsPage(boards, connections).Render(c.Request().Context(), c.Response().Writer)
}

// GetBoard shows a specific board
func (h *BoardHandler) GetBoard(c echo.Context) error {
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid board ID")
	}

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	board, err := svc.GetBoardWithData(boardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
	}

	return templates.BoardPage(board).Render(c.Request().Context(), c.Response().Writer)
}

type CreateBoardRequest struct {
	Name             string `form:"name"`
	DbType           string `form:"db_type"`
	PgConnectionID   int64  `form:"pg_connection_id"`
	PgDatabaseName   string `form:"pg_database_name"`
}

// CreateBoard creates a new board
func (h *BoardHandler) CreateBoard(c echo.Context) error {
	var req CreateBoardRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	req.Name = validation.SanitizeName(req.Name)
	if req.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}

	if req.DbType == "" {
		req.DbType = "local"
	}

	var pgConnID *int64
	if req.DbType == "postgres" && req.PgConnectionID > 0 {
		pgConnID = &req.PgConnectionID
	}

	_, err := h.bm.CreateBoard(req.Name, req.DbType, pgConnID, req.PgDatabaseName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create board: "+err.Error())
	}

	boards, err := h.bm.GetAllBoards()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load boards")
	}

	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	return templates.BoardsList(boards, connections).Render(c.Request().Context(), c.Response().Writer)
}

type RenameBoardRequest struct {
	Name string `form:"name"`
}

// RenameBoard renames a board
func (h *BoardHandler) RenameBoard(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid board ID")
	}

	var req RenameBoardRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	req.Name = validation.SanitizeName(req.Name)
	if req.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}

	if err := h.bm.RenameBoard(id, req.Name); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to rename board")
	}

	boards, err := h.bm.GetAllBoards()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load boards")
	}

	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	return templates.BoardsList(boards, connections).Render(c.Request().Context(), c.Response().Writer)
}

// DeleteBoard deletes a board
func (h *BoardHandler) DeleteBoard(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid board ID")
	}

	if err := h.bm.DeleteBoard(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete board")
	}

	boards, err := h.bm.GetAllBoards()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load boards")
	}

	connections, err := h.bm.PgConnRepo().GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load connections")
	}

	return templates.BoardsList(boards, connections).Render(c.Request().Context(), c.Response().Writer)
}
