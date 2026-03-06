package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"krizzy/internal/models"
	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
)

type RealtimeHandler struct {
	bm  *services.BoardManager
	hub *services.BoardEventHub
}

func NewRealtimeHandler(bm *services.BoardManager, hub *services.BoardEventHub) *RealtimeHandler {
	return &RealtimeHandler{bm: bm, hub: hub}
}

func (h *RealtimeHandler) StreamBoardEvents(c echo.Context) error {
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid board ID")
	}

	if _, err := h.bm.GetServiceForBoard(boardID); err != nil {
		return c.String(http.StatusNotFound, "Board not found")
	}

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, "text/event-stream")
	res.Header().Set(echo.HeaderCacheControl, "no-cache")
	res.Header().Set(echo.HeaderConnection, "keep-alive")
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)
	res.Flush()

	ch := h.hub.Subscribe(boardID)
	defer h.hub.Unsubscribe(boardID, ch)

	if _, err := fmt.Fprint(res, ": connected\n\n"); err != nil {
		return nil
	}
	res.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := fmt.Fprint(res, ": keepalive\n\n"); err != nil {
				return nil
			}
			res.Flush()
		case event, ok := <-ch:
			if !ok {
				return nil
			}

			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}

			if _, err := fmt.Fprintf(res, "event: board-update\ndata: %s\n\n", payload); err != nil {
				return nil
			}
			res.Flush()
		}
	}
}

func (h *RealtimeHandler) GetColumnsContainer(c echo.Context) error {
	board, err := h.loadBoard(c)
	if err != nil {
		return err
	}

	return templates.ColumnsContainer(board).Render(c.Request().Context(), c.Response().Writer)
}

func (h *RealtimeHandler) GetColumn(c echo.Context) error {
	boardID, columnID, svc, err := h.loadBoardAndColumn(c)
	if err != nil {
		return err
	}

	column, err := svc.ColumnRepo.GetByID(columnID)
	if err != nil {
		return c.String(http.StatusNotFound, "Column not found")
	}
	if column.BoardID != boardID {
		return c.String(http.StatusNotFound, "Column not found")
	}

	board, err := svc.GetBoardWithData(boardID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load board")
	}

	for i := range board.Columns {
		if board.Columns[i].ID == columnID {
			return templates.ColumnComponent(&board.Columns[i], boardID).Render(c.Request().Context(), c.Response().Writer)
		}
	}

	return c.String(http.StatusNotFound, "Column not found")
}

func (h *RealtimeHandler) GetCard(c echo.Context) error {
	boardID, cardID, svc, err := h.loadBoardAndCard(c)
	if err != nil {
		return err
	}

	card, err := svc.GetCardWithDetails(cardID)
	if err != nil {
		return c.String(http.StatusNotFound, "Card not found")
	}

	column, err := svc.ColumnRepo.GetByID(card.ColumnID)
	if err != nil || column.BoardID != boardID {
		return c.String(http.StatusNotFound, "Card not found")
	}

	return templates.CardComponent(card, boardID).Render(c.Request().Context(), c.Response().Writer)
}

func (h *RealtimeHandler) loadBoard(c echo.Context) (*models.Board, error) {
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, c.String(http.StatusBadRequest, "Invalid board ID")
	}

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return nil, c.String(http.StatusNotFound, "Board not found")
	}

	board, err := svc.GetBoardWithData(boardID)
	if err != nil {
		return nil, c.String(http.StatusInternalServerError, "Failed to load board")
	}

	return board, nil
}

func (h *RealtimeHandler) loadBoardAndColumn(c echo.Context) (int64, int64, *services.KanbanService, error) {
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, 0, nil, c.String(http.StatusBadRequest, "Invalid board ID")
	}

	columnID, err := strconv.ParseInt(c.Param("columnId"), 10, 64)
	if err != nil {
		return 0, 0, nil, c.String(http.StatusBadRequest, "Invalid column ID")
	}

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return 0, 0, nil, c.String(http.StatusNotFound, "Board not found")
	}

	return boardID, columnID, svc, nil
}

func (h *RealtimeHandler) loadBoardAndCard(c echo.Context) (int64, int64, *services.KanbanService, error) {
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, 0, nil, c.String(http.StatusBadRequest, "Invalid board ID")
	}

	cardID, err := strconv.ParseInt(c.Param("cardId"), 10, 64)
	if err != nil {
		return 0, 0, nil, c.String(http.StatusBadRequest, "Invalid card ID")
	}

	svc, err := h.bm.GetServiceForBoard(boardID)
	if err != nil {
		return 0, 0, nil, c.String(http.StatusNotFound, "Board not found")
	}

	return boardID, cardID, svc, nil
}

func publishBoardEvent(hub *services.BoardEventHub, event services.BoardEvent) {
	if hub == nil {
		return
	}
	hub.Publish(event)
}

func requestClientID(c echo.Context) string {
	return c.Request().Header.Get("X-Client-ID")
}
