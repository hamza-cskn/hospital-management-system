package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hamza/proglabodev3/api/config"
	"github.com/hamza/proglabodev3/api/handlers"
	"github.com/hamza/proglabodev3/api/middleware"
	"github.com/hamza/proglabodev3/api/models"
)

func createRootAccount(db *config.Database) error {
	rootUser := models.User{
		Email:     "root@root.com",
		Password:  "root",
		FirstName: "Root",
		LastName:  "Root",
		Role:      models.RoleAdmin,
	}

	_, err := handlers.CreateUser(rootUser, db)
	if err != nil && err.Error() != "email already exists" {
		return err
	}
	return nil
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	createRootAccount(db)

	// Initialize Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Frontend routes
	frontend := router.Group("")
	{
		// Public pages
		frontend.GET("/", handlers.GetStaticPageHandler("index.html"))
		frontend.GET("/login", handlers.GetStaticPageHandler("login.html"))
		frontend.GET("/register", handlers.GetStaticPageHandler("register.html"))
		frontend.GET("/js/core.js", handlers.GetStaticPageHandler("js/core.js"))
		frontend.GET("/css/styles.css", handlers.GetStaticPageHandler("css/styles.css"))

		frontend.GET("/profile", handlers.GetStaticPageHandler("profile.html"))
		frontend.GET("/patient-dashboard", handlers.GetStaticPageHandler("patient-dashboard.html"))
		frontend.GET("/doctor-dashboard", handlers.GetStaticPageHandler("doctor-dashboard.html"))
		frontend.GET("/admin-dashboard", handlers.GetStaticPageHandler("admin-dashboard.html"))
		frontend.GET("/book-appointment", handlers.GetStaticPageHandler("book-appointment.html"))
		frontend.GET("/create-doctor", handlers.GetStaticPageHandler("create-doctor.html"))
	}

	// API routes
	api := router.Group("/api")
	{
		// Public routes
		api.POST("/auth/register", handlers.Register(db))
		api.POST("/auth/login", handlers.Login(db, cfg))

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			protected.GET("/users/profile", handlers.GetUserProfile(db))
			protected.PUT("/users/profile", handlers.UpdateUserProfile(db))

			// Appointment routes
			protected.POST("/appointments", handlers.CreateAppointment(db))
			protected.GET("/appointments", handlers.GetAppointments(db))
			protected.GET("/appointments/:id", handlers.GetAppointment(db))
			protected.PUT("/appointments/:id", handlers.UpdateAppointment(db))
			protected.DELETE("/appointments/:id", handlers.DeleteAppointment(db))
			protected.GET("/users", handlers.GetAllUsers(db))

			// Doctor routes
			doctorRoutes := protected.Group("/doctors")
			doctorRoutes.Use(middleware.RequireRole(string(models.RoleDoctor), string(models.RoleAdmin)))
			{
				//doctorRoutes.GET("/appointments", handlers.GetDoctorAppointments(db))
				doctorRoutes.PUT("/expertises", handlers.UpdateDoctorExpertise(db))
				doctorRoutes.GET("/expertises", handlers.GetAllExpertises(db))
			}

			// Admin routes
			adminRoutes := protected.Group("/admin")
			adminRoutes.Use(middleware.RequireRole(string(models.RoleAdmin)))
			{
				adminRoutes.POST("/doctors", handlers.CreateDoctor(db))
				adminRoutes.POST("/expertises", handlers.CreateExpertise(db))
				adminRoutes.PUT("/users/:id", handlers.UpdateUser(db))
				adminRoutes.DELETE("/users/:id", handlers.DeleteUser(db))
			}
		}
	}
	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
