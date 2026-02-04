package handlers

import (
	"net/http"

	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type BoardHandler struct {
	service *services.KanbanService
}

func NewBoardHandler(service *services.KanbanService) *BoardHandler {
	return &BoardHandler{service: service}
}

func (h *BoardHandler) GetBoard(c echo.Context) error {
	board, err := h.service.EnsureDefaultBoard()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	// If HTMX request, return just the board content
	if c.Request().Header.Get("HX-Request") == "true" {
		return templates.BoardContent(board).Render(c.Request().Context(), c.Response().Writer)
	}

	return templates.BoardPage(board).Render(c.Request().Context(), c.Response().Writer)
}
