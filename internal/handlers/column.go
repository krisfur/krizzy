package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/models"
	"krizzy/internal/repository"
	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ColumnHandler struct {
	service    *services.KanbanService
	columnRepo repository.ColumnRepository
}

func NewColumnHandler(service *services.KanbanService, columnRepo repository.ColumnRepository) *ColumnHandler {
	return &ColumnHandler{
		service:    service,
		columnRepo: columnRepo,
	}
}

type CreateColumnRequest struct {
	Name    string `form:"name"`
	BoardID int64  `form:"board_id"`
}

func (h *ColumnHandler) CreateColumn(c echo.Context) error {
	var req CreateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	column := &models.Column{
		BoardID: req.BoardID,
		Name:    req.Name,
	}

	if err := h.columnRepo.Create(column); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create column")
	}

	// Return the board partial to refresh the view
	board, err := h.service.GetBoardWithData(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateColumnRequest struct {
	Name         string `form:"name"`
	IsDoneColumn bool   `form:"is_done_column"`
}

func (h *ColumnHandler) UpdateColumn(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid column ID")
	}

	var req UpdateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	column, err := h.columnRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Column not found")
	}

	column.Name = req.Name
	column.IsDoneColumn = req.IsDoneColumn

	if err := h.columnRepo.Update(column); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update column")
	}

	board, err := h.service.GetBoardWithData(column.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ColumnHandler) DeleteColumn(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid column ID")
	}

	column, err := h.columnRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Column not found")
	}

	boardID := column.BoardID

	if err := h.columnRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete column")
	}

	board, err := h.service.GetBoardWithData(boardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type ReorderColumnsRequest struct {
	BoardID   int64   `form:"board_id"`
	ColumnIDs []int64 `form:"column_ids"`
}

func (h *ColumnHandler) ReorderColumns(c echo.Context) error {
	var req ReorderColumnsRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if err := h.columnRepo.Reorder(req.BoardID, req.ColumnIDs); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to reorder columns")
	}

	return c.NoContent(http.StatusOK)
}
