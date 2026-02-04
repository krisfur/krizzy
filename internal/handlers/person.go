package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/models"
	"krizzy/internal/repository"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type PersonHandler struct {
	personRepo repository.PersonRepository
}

func NewPersonHandler(personRepo repository.PersonRepository) *PersonHandler {
	return &PersonHandler{personRepo: personRepo}
}

func (h *PersonHandler) ListPeople(c echo.Context) error {
	people, err := h.personRepo.GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.PeopleList(people).Render(c.Request().Context(), c.Response().Writer)
}

type CreatePersonRequest struct {
	Name string `form:"name"`
}

func (h *PersonHandler) CreatePerson(c echo.Context) error {
	var req CreatePersonRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if req.Name == "" {
		return c.String(http.StatusBadRequest, "Name is required")
	}

	person := &models.Person{Name: req.Name}
	if err := h.personRepo.Create(person); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create person")
	}

	people, err := h.personRepo.GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.PeopleList(people).Render(c.Request().Context(), c.Response().Writer)
}

func (h *PersonHandler) DeletePerson(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid person ID")
	}

	if err := h.personRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete person")
	}

	people, err := h.personRepo.GetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.PeopleList(people).Render(c.Request().Context(), c.Response().Writer)
}
