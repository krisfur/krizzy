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

type CardHandler struct {
	bm *services.BoardManager
}

func NewCardHandler(bm *services.BoardManager) *CardHandler {
	return &CardHandler{bm: bm}
}

type CreateCardRequest struct {
	ColumnID int64  `form:"column_id"`
	Title    string `form:"title"`
	BoardID  int64  `form:"board_id"`
}

func (h *CardHandler) CreateCard(c echo.Context) error {
	var req CreateCardRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	req.Title = validation.SanitizeName(req.Title)

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	card := &models.Card{
		ColumnID: req.ColumnID,
		Title:    req.Title,
	}

	if err := svc.CardRepo.Create(card); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create card")
	}

	board, err := svc.GetBoardWithData(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type UpdateCardRequest struct {
	Title       string `form:"title"`
	Description string `form:"description"`
	BoardID     int64  `form:"board_id"`
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

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	card, err := svc.CardRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	card.Title = req.Title
	card.Description = req.Description

	if err := svc.CardRepo.Update(card); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update card")
	}

	cardWithDetails, err := svc.GetCardWithDetails(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load card")
	}

	people, err := svc.PersonRepo.GetByBoardID(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.CardModal(cardWithDetails, people, req.BoardID).Render(c.Request().Context(), c.Response().Writer)
}

func (h *CardHandler) DeleteCard(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	boardID, _ := strconv.ParseInt(c.QueryParam("board_id"), 10, 64)

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.CardRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete card")
	}

	board, err := svc.GetBoardWithData(boardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
}

type MoveCardRequest struct {
	ColumnID int64 `json:"column_id" form:"column_id"`
	Position int   `json:"position" form:"position"`
	BoardID  int64 `json:"board_id" form:"board_id"`
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

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.MoveCard(id, req.ColumnID, req.Position); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to move card")
	}

	return c.NoContent(http.StatusOK)
}

type UpdateAssigneesRequest struct {
	PersonIDs []int64 `form:"person_ids"`
	BoardID   int64   `form:"board_id"`
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

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.PersonRepo.SetCardAssignees(id, req.PersonIDs); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to update assignees")
	}

	cardWithDetails, err := svc.GetCardWithDetails(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load card")
	}

	people, err := svc.PersonRepo.GetByBoardID(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.AssigneePicker(cardWithDetails.ID, cardWithDetails.Assignees, people, req.BoardID).Render(c.Request().Context(), c.Response().Writer)
}
