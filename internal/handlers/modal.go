package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ModalHandler struct {
	bm *services.BoardManager
}

func NewModalHandler(bm *services.BoardManager) *ModalHandler {
	return &ModalHandler{bm: bm}
}

func (h *ModalHandler) GetCardModal(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	boardID, _ := strconv.ParseInt(c.QueryParam("board_id"), 10, 64)

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	card, err := svc.GetCardWithDetails(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	people, err := svc.PersonRepo.GetByBoardID(boardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.CardModal(card, people, boardID).Render(c.Request().Context(), c.Response().Writer)
}
