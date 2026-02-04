package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/repository"
	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type ModalHandler struct {
	service    *services.KanbanService
	personRepo repository.PersonRepository
}

func NewModalHandler(service *services.KanbanService, personRepo repository.PersonRepository) *ModalHandler {
	return &ModalHandler{
		service:    service,
		personRepo: personRepo,
	}
}

func (h *ModalHandler) GetCardModal(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	card, err := h.service.GetCardWithDetails(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	people, err := h.personRepo.GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.CardModal(card, people).Render(c.Request().Context(), c.Response().Writer)
}
