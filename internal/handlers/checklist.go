package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/models"
	"krizzy/internal/repository"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ChecklistHandler struct {
	checklistRepo repository.ChecklistRepository
}

func NewChecklistHandler(checklistRepo repository.ChecklistRepository) *ChecklistHandler {
	return &ChecklistHandler{checklistRepo: checklistRepo}
}

type CreateChecklistItemRequest struct {
	Content string `form:"content"`
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

	item := &models.ChecklistItem{
		CardID:  cardID,
		Content: req.Content,
	}

	if err := h.checklistRepo.Create(item); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create checklist item")
	}

	items, err := h.checklistRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load checklist")
	}

	return templates.ChecklistComponent(cardID, items).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateChecklistItemRequest struct {
	Content     string `form:"content"`
	IsCompleted bool   `form:"is_completed"`
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

	item, err := h.checklistRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}

	if req.Content != "" {
		item.Content = req.Content
	}
	item.IsCompleted = req.IsCompleted

	if err := h.checklistRepo.Update(item); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update item")
	}

	items, err := h.checklistRepo.GetByCardID(item.CardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load checklist")
	}

	return templates.ChecklistComponent(item.CardID, items).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ChecklistHandler) DeleteItem(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid item ID")
	}

	item, err := h.checklistRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}

	cardID := item.CardID

	if err := h.checklistRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete item")
	}

	items, err := h.checklistRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load checklist")
	}

	return templates.ChecklistComponent(cardID, items).Render(c.Request().Context(), c.Response().Writer)
}

type ReorderChecklistRequest struct {
	ItemIDs []int64 `form:"item_ids"`
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

	if err := h.checklistRepo.Reorder(cardID, req.ItemIDs); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to reorder checklist")
	}

	return c.NoContent(http.StatusOK)
}
