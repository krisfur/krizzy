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

type CommentHandler struct {
	service     *services.KanbanService
	commentRepo repository.CommentRepository
}

func NewCommentHandler(service *services.KanbanService, commentRepo repository.CommentRepository) *CommentHandler {
	return &CommentHandler{
		service:     service,
		commentRepo: commentRepo,
	}
}

type CreateCommentRequest struct {
	Content string `form:"content"`
}

func (h *CommentHandler) CreateComment(c echo.Context) error {
	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid card ID")
	}

	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if req.Content == "" {
		return c.String(http.StatusBadRequest, "Content is required")
	}

	comment := &models.Comment{
		CardID:  cardID,
		Content: req.Content,
	}

	if err := h.commentRepo.Create(comment); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create comment")
	}

	comments, err := h.commentRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load comments")
	}

	return templates.CommentsList(cardID, comments).Render(c.Request().Context(), c.Response().Writer)
}

func (h *CommentHandler) DeleteComment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid comment ID")
	}

	comment, err := h.commentRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Comment not found")
	}

	cardID := comment.CardID

	if err := h.commentRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete comment")
	}

	comments, err := h.commentRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load comments")
	}

	return templates.CommentsList(cardID, comments).Render(c.Request().Context(), c.Response().Writer)
}
