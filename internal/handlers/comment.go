package handlers

import (
	"net/http"
	"strconv"

	"krizzy/internal/models"
	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	bm *services.BoardManager
}

func NewCommentHandler(bm *services.BoardManager) *CommentHandler {
	return &CommentHandler{bm: bm}
}

type CreateCommentRequest struct {
	Content string `form:"content"`
	BoardID int64  `form:"board_id"`
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

	svc, err := h.bm.GetServiceForBoard(req.BoardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	comment := &models.Comment{
		CardID:  cardID,
		Content: req.Content,
	}

	if err := svc.CommentRepo.Create(comment); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create comment")
	}

	comments, err := svc.CommentRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load comments")
	}

	return templates.CommentsList(cardID, comments, req.BoardID).Render(c.Request().Context(), c.Response().Writer)
}

func (h *CommentHandler) DeleteComment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid comment ID")
	}

	boardID, _ := strconv.ParseInt(c.QueryParam("board_id"), 10, 64)

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	comment, err := svc.CommentRepo.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Comment not found")
	}

	cardID := comment.CardID

	if err := svc.CommentRepo.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete comment")
	}

	comments, err := svc.CommentRepo.GetByCardID(cardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load comments")
	}

	return templates.CommentsList(cardID, comments, boardID).Render(c.Request().Context(), c.Response().Writer)
}
