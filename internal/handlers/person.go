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

type PersonHandler struct {
	bm *services.BoardManager
}

func NewPersonHandler(bm *services.BoardManager) *PersonHandler {
	return &PersonHandler{bm: bm}
}

type CreatePersonRequest struct {
	Name    string `form:"name"`
	BoardID int64  `form:"board_id"`
}

func (h *PersonHandler) CreatePerson(c echo.Context) error {
	var req CreatePersonRequest
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

	person := &models.Person{Name: req.Name, BoardID: req.BoardID}
	if err := svc.PersonRepo.Create(person); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create person")
	}

	people, err := svc.PersonRepo.GetByBoardID(req.BoardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.PeopleList(people, req.BoardID).Render(c.Request().Context(), c.Response().Writer)
}

type DeletePersonRequest struct {
	BoardID int64 `query:"board_id"`
}

func (h *PersonHandler) DeletePerson(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid person ID")
	}

	boardID, _ := strconv.ParseInt(c.QueryParam("board_id"), 10, 64)

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	if err := svc.PersonRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete person")
	}

	people, err := svc.PersonRepo.GetByBoardID(boardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load people")
	}

	return templates.PeopleList(people, boardID).Render(c.Request().Context(), c.Response().Writer)
}
