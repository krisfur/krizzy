package main

import (
	"log"
	"net/http"
	"strconv"

	"krizzy/internal/config"
	"krizzy/internal/database"
	"krizzy/internal/handlers"
	"krizzy/internal/repository"
	"krizzy/internal/services"
	"krizzy/templates"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories (always local SQLite for metadata)
	boardRepo := repository.NewSQLiteBoardRepository(db.DB())
	pgConnRepo := repository.NewSQLitePgConnectionRepository(db.DB())

	// Initialize BoardManager
	bm := services.NewBoardManager(db, boardRepo, pgConnRepo)
	defer bm.Close()

	// Initialize handlers
	boardHandler := handlers.NewBoardHandler(bm)
	columnHandler := handlers.NewColumnHandler(bm)
	cardHandler := handlers.NewCardHandler(bm)
	modalHandler := handlers.NewModalHandler(bm)
	personHandler := handlers.NewPersonHandler(bm)
	commentHandler := handlers.NewCommentHandler(bm)
	checklistHandler := handlers.NewChecklistHandler(bm)
	connectionHandler := handlers.NewConnectionHandler(bm)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files
	e.Static("/static", "static")

	// Board list routes
	e.GET("/", boardHandler.ListBoards)
	e.POST("/boards", boardHandler.CreateBoard)
	e.GET("/boards/:id", boardHandler.GetBoard)
	e.PUT("/boards/:id", boardHandler.RenameBoard)
	e.DELETE("/boards/:id", boardHandler.DeleteBoard)

	// Board-scoped people modal
	e.GET("/boards/:id/people", func(c echo.Context) error {
		boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid board ID")
		}
		svc, err := bm.GetServiceForBoard(boardID)
		if err != nil {
			return c.String(http.StatusNotFound, "Board not found")
		}
		people, err := svc.PersonRepo.GetByBoardID(boardID)
		if err != nil {
			return err
		}
		return templates.PeopleModal(people, boardID).Render(c.Request().Context(), c.Response().Writer)
	})

	// Column routes
	e.POST("/columns", columnHandler.CreateColumn)
	e.PUT("/columns/:id", columnHandler.UpdateColumn)
	e.DELETE("/columns/:id", columnHandler.DeleteColumn)
	e.POST("/columns/reorder", columnHandler.ReorderColumns)

	// Card routes
	e.POST("/cards", cardHandler.CreateCard)
	e.GET("/cards/:id/modal", modalHandler.GetCardModal)
	e.PUT("/cards/:id", cardHandler.UpdateCard)
	e.DELETE("/cards/:id", cardHandler.DeleteCard)
	e.POST("/cards/:id/move", cardHandler.MoveCard)
	e.POST("/cards/:id/assignees", cardHandler.UpdateAssignees)

	// Comment routes
	e.POST("/cards/:id/comments", commentHandler.CreateComment)
	e.DELETE("/comments/:id", commentHandler.DeleteComment)

	// Checklist routes
	e.POST("/cards/:id/checklist", checklistHandler.CreateItem)
	e.PUT("/checklist/:id", checklistHandler.UpdateItem)
	e.DELETE("/checklist/:id", checklistHandler.DeleteItem)
	e.POST("/cards/:id/checklist/reorder", checklistHandler.ReorderItems)

	// People routes
	e.POST("/people", personHandler.CreatePerson)
	e.DELETE("/people/:id", personHandler.DeletePerson)

	// Connection routes
	e.GET("/connections", connectionHandler.ListConnections)
	e.POST("/connections", connectionHandler.CreateConnection)
	e.POST("/connections/:id/test", connectionHandler.TestConnection)
	e.DELETE("/connections/:id", connectionHandler.DeleteConnection)

	// Start server
	log.Printf("Starting Krizzy on %s", cfg.ServerAddress)
	log.Fatal(e.Start(cfg.ServerAddress))
}
