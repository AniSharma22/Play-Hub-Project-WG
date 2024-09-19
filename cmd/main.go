package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"project2/internal/api/handlers"
	"project2/internal/api/routes"
	"project2/internal/app/repositories"
	"project2/internal/app/services"
	"project2/internal/config"
	"project2/internal/db"
	"project2/pkg/utils"
	"syscall"
)

func main() {

	// Initialize PostgresSQL client
	client, err := db.PostgresInitClient()
	if err != nil {
		log.Fatal("Error initializing PostgresSQL client:", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatal("Error closing PostgresSQL client:", err)
		}
	}()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Graceful shutdown initiated...")
		client.Close()
		fmt.Println("All operations completed. Exiting.")
		os.Exit(0)
	}()

	// Initialize repositories
	userRepo := repositories.NewUserRepo(client)
	gameRepo := repositories.NewGameRepo(client)
	slotRepo := repositories.NewSlotRepo(client)
	invitationRepo := repositories.NewInvitationRepo(client)
	bookingRepo := repositories.NewBookingRepo(client)
	leaderboardRepo := repositories.NewLeaderboardRepo(client)
	notificationRepo := repositories.NewNotificationRepo(client)

	// Initialize services
	gameService := services.NewGameService(gameRepo)
	slotService := services.NewSlotService(slotRepo)
	userService := services.NewUserService(userRepo)
	notificationService := services.NewNotificationService(notificationRepo)
	authService := services.NewAuthService(userRepo, userService)
	bookingService := services.NewBookingService(bookingRepo, slotService, gameService)
	invitationService := services.NewInvitationService(invitationRepo, bookingService, slotService)
	leaderboardService := services.NewLeaderboardService(leaderboardRepo, bookingService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	gameHandler := handlers.NewGameHandler(gameService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	slotHandler := handlers.NewSlotHandler(slotService)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Initialize Router and handlers
	r := mux.NewRouter()
	routes.InitialiseUserRouter(r, userHandler)
	routes.InitialiseBookingRouter(r, bookingHandler)
	routes.InitialiseAuthRouter(r, authHandler)
	routes.InitialiseGameRouter(r, gameHandler)
	routes.InitialiseSlotRouter(r, slotHandler)
	routes.InitialiseInvitationRouter(r, invitationHandler)
	routes.InitialiseLeaderboardRouter(r, leaderboardHandler)
	routes.InitialiseNotificationRouter(r, notificationHandler)

	// todo: have to automate this process
	// Insert today's slots
	err = utils.InsertAllSlots(context.Background(), slotRepo, gameRepo)
	if err != nil {
		log.Fatal("Error inserting slots:", err)
	}

	http.Handle("/", r)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Healthy"))
		w.WriteHeader(http.StatusOK)
	})
	fmt.Println("api is running good")
	log.Fatal(http.ListenAndServe(config.PORT, nil))

}
