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

type ColumnHandler struct {
	bm *services.BoardManager
}

func NewColumnHandler(bm *services.BoardManager) *ColumnHandler {
	return &ColumnHandler{bm: bm}
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

	req.Name = validation.SanitizeName(req.Name)

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	column := &models.Column{
		BoardID: req.BoardID,
		Name:    req.Name,
	}

	if err := svc.ColumnRepo.Create(column); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create column")
	}

	board, err := svc.GetBoardWithData(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateColumnRequest struct {
	Name         string `form:"name"`
	IsDoneColumn bool   `form:"is_done_column"`
	BoardID      int64  `form:"board_id"`
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

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	column, err := svc.ColumnRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Column not found")
	}

	column.Name = req.Name
	column.IsDoneColumn = req.IsDoneColumn

	if err := svc.ColumnRepo.Update(column); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update column")
	}

	board, err := svc.GetBoardWithData(req.BoardID)
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

	boardID, _ := strconv.ParseInt(c.QueryParam("board_id"), 10, 64)

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.ColumnRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete column")
	}

	board, err := svc.GetBoardWithData(boardID)
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

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.ColumnRepo.Reorder(req.BoardID, req.ColumnIDs); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to reorder columns")
	}

	return c.NoContent(http.StatusOK)
}
