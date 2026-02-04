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

type CardHandler struct {
	service    *services.KanbanService
	cardRepo   repository.CardRepository
	columnRepo repository.ColumnRepository
	personRepo repository.PersonRepository
}

func NewCardHandler(
	service *services.KanbanService,
	cardRepo repository.CardRepository,
	columnRepo repository.ColumnRepository,
	personRepo repository.PersonRepository,
) *CardHandler {
	return &CardHandler{
		service:    service,
		cardRepo:   cardRepo,
		columnRepo: columnRepo,
		personRepo: personRepo,
	}
}

type CreateCardRequest struct {
	ColumnID int64  `form:"column_id"`
	Title    string `form:"title"`
}

func (h *CardHandler) CreateCard(c echo.Context) error {
	var req CreateCardRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	card := &models.Card{
		ColumnID: req.ColumnID,
		Title:    req.Title,
	}

	if err := h.cardRepo.Create(card); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create card")
	}

	column, err := h.columnRepo.GetByID(req.ColumnID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load column")
	}

	board, err := h.service.GetBoardWithData(column.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateCardRequest struct {
	Title       string `form:"title"`
	Description string `form:"description"`
}

func (h *CardHandler) UpdateCard(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	var req UpdateCardRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	card, err := h.cardRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	card.Title = req.Title
	card.Description = req.Description

	if err := h.cardRepo.Update(card); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update card")
	}

	// Return updated modal
	cardWithDetails, err := h.service.GetCardWithDetails(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load card")
	}

	people, err := h.personRepo.GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.CardModal(cardWithDetails, people).Render(c.Request().Context(), c.Response().Writer)
}

func (h *CardHandler) DeleteCard(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	card, err := h.cardRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	column, err := h.columnRepo.GetByID(card.ColumnID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load column")
	}

	if err := h.cardRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete card")
	}

	board, err := h.service.GetBoardWithData(column.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type MoveCardRequest struct {
	ColumnID int64 `json:"column_id" form:"column_id"`
	Position int   `json:"position" form:"position"`
}

func (h *CardHandler) MoveCard(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	var req MoveCardRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if err := h.service.MoveCard(id, req.ColumnID, req.Position); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to move card")
	}

	return c.NoContent(http.StatusOK)
}

type UpdateAssigneesRequest struct {
	PersonIDs []int64 `form:"person_ids"`
}

func (h *CardHandler) UpdateAssignees(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	var req UpdateAssigneesRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if err := h.personRepo.SetCardAssignees(id, req.PersonIDs); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update assignees")
	}

	// Return updated assignee picker
	cardWithDetails, err := h.service.GetCardWithDetails(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load card")
	}

	people, err := h.personRepo.GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.AssigneePicker(cardWithDetails.ID, cardWithDetails.Assignees, people).Render(c.Request().Context(), c.Response().Writer)
}
