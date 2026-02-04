package main

import (
	"log"

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

	// Initialize repositories
	boardRepo := repository.NewSQLiteBoardRepository(db.DB())
	columnRepo := repository.NewSQLiteColumnRepository(db.DB())
	cardRepo := repository.NewSQLiteCardRepository(db.DB())
	personRepo := repository.NewSQLitePersonRepository(db.DB())
	commentRepo := repository.NewSQLiteCommentRepository(db.DB())
	checklistRepo := repository.NewSQLiteChecklistRepository(db.DB())

	// Initialize service
	kanbanService := services.NewKanbanService(
		boardRepo,
		columnRepo,
		cardRepo,
		personRepo,
		commentRepo,
		checklistRepo,
	)

	// Ensure default board exists
	if _, err := kanbanService.EnsureDefaultBoard(); err != nil {
		log.Fatalf("Failed to ensure default board: %v", err)
	}

	// Initialize handlers
	boardHandler := handlers.NewBoardHandler(kanbanService)
	columnHandler := handlers.NewColumnHandler(kanbanService, columnRepo)
	cardHandler := handlers.NewCardHandler(kanbanService, cardRepo, columnRepo, personRepo)
	modalHandler := handlers.NewModalHandler(kanbanService, personRepo)
	personHandler := handlers.NewPersonHandler(personRepo)
	commentHandler := handlers.NewCommentHandler(kanbanService, commentRepo)
	checklistHandler := handlers.NewChecklistHandler(checklistRepo)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files
	e.Static("/static", "static")

	// Routes
	e.GET("/", boardHandler.GetBoard)

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
	e.GET("/people", func(c echo.Context) error {
		people, err := personRepo.GetAll()
		if err != nil {
			return err
		}
		return templates.PeopleModal(people).Render(c.Request().Context(), c.Response().Writer)
	})
	e.POST("/people", personHandler.CreatePerson)
	e.DELETE("/people/:id", personHandler.DeletePerson)

	// Start server
	log.Printf("Starting Krizzy on %s", cfg.ServerAddress)
	log.Fatal(e.Start(cfg.ServerAddress))
}
