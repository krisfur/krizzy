package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"krizzy/internal/models"
	"krizzy/internal/services"
	"krizzy/internal/validation"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ColumnHandler struct {
	bm  *services.BoardManager
	hub *services.BoardEventHub
}

func NewColumnHandler(bm *services.BoardManager, hub *services.BoardEventHub) *ColumnHandler {
	return &ColumnHandler{bm: bm, hub: hub}
}

type CreateColumnRequest struct {
	Name    string `form:"name"`
	BoardID int64  `form:"board_id"`
}

func isDoneColumnName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "done")
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
		BoardID:      req.BoardID,
		Name:         req.Name,
		IsDoneColumn: isDoneColumnName(req.Name),
	}

	if err := svc.ColumnRepo.Create(column); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create column")
	}

	publishBoardEvent(h.hub, services.BoardEvent{
		Type:     "column.created",
		BoardID:  req.BoardID,
		ColumnID: column.ID,
		ClientID: requestClientID(c),
	})

	board, err := svc.GetBoardWithData(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateColumnRequest struct {
	Name    string `form:"name"`
	BoardID int64  `form:"board_id"`
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

	req.Name = validation.SanitizeName(req.Name)
	if req.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
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
	column.IsDoneColumn = isDoneColumnName(req.Name)

	if err := svc.ColumnRepo.Update(column); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update column")
	}

	publishBoardEvent(h.hub, services.BoardEvent{
		Type:     "column.updated",
		BoardID:  req.BoardID,
		ColumnID: column.ID,
		ClientID: requestClientID(c),
	})

	board, err := svc.GetBoardWithData(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	for i := range board.Columns {
		if board.Columns[i].ID == column.ID {
			return templates.ColumnComponent(&board.Columns[i], req.BoardID).Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return c.String(http.StatusNotFound, "Column not found")
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

	column, err := svc.ColumnRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Column not found")
	}

	if err := svc.ColumnRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete column")
	}

	publishBoardEvent(h.hub, services.BoardEvent{
		Type:     "column.deleted",
		BoardID:  boardID,
		ColumnID: column.ID,
		ClientID: requestClientID(c),
	})

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

	publishBoardEvent(h.hub, services.BoardEvent{
		Type:     "column.reordered",
		BoardID:  req.BoardID,
		ClientID: requestClientID(c),
	})

	return c.NoContent(http.StatusOK)
}
