package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/models"
	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ChecklistHandler struct {
	bm *services.BoardManager
}

func NewChecklistHandler(bm *services.BoardManager) *ChecklistHandler {
	return &ChecklistHandler{bm: bm}
}

type CreateChecklistItemRequest struct {
	Content string `form:"content"`
	BoardID int64  `form:"board_id"`
}

func (h *ChecklistHandler) CreateItem(c echo.Context) error {
	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	var req CreateChecklistItemRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if req.Content == "" {
		return c.String(http.StatusBadRequest, "Content is required")
	}

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	item := &models.ChecklistItem{
		CardID:  cardID,
		Content: req.Content,
	}

	if err := svc.ChecklistRepo.Create(item); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create checklist item")
	}

	items, err := svc.ChecklistRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load checklist")
	}

	return templates.ChecklistComponent(cardID, items, req.BoardID).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateChecklistItemRequest struct {
	Content     string `form:"content"`
	IsCompleted bool   `form:"is_completed"`
	BoardID     int64  `form:"board_id"`
}

func (h *ChecklistHandler) UpdateItem(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid item ID")
	}

	var req UpdateChecklistItemRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	item, err := svc.ChecklistRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}

	if req.Content != "" {
		item.Content = req.Content
	}
	item.IsCompleted = req.IsCompleted

	if err := svc.ChecklistRepo.Update(item); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update item")
	}

	items, err := svc.ChecklistRepo.GetByCardID(item.CardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load checklist")
	}

	return templates.ChecklistComponent(item.CardID, items, req.BoardID).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ChecklistHandler) DeleteItem(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid item ID")
	}

	boardID, _ := strconv.ParseInt(c.QueryParam("board_id"), 10, 64)

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	item, err := svc.ChecklistRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}

	cardID := item.CardID

	if err := svc.ChecklistRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete item")
	}

	items, err := svc.ChecklistRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load checklist")
	}

	return templates.ChecklistComponent(cardID, items, boardID).Render(c.Request().Context(), c.Response().Writer)
}

type ReorderChecklistRequest struct {
	ItemIDs []int64 `form:"item_ids"`
	BoardID int64   `form:"board_id"`
}

func (h *ChecklistHandler) ReorderItems(c echo.Context) error {
	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	var req ReorderChecklistRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.ChecklistRepo.Reorder(cardID, req.ItemIDs); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to reorder checklist")
	}

	return c.NoContent(http.StatusOK)
}
