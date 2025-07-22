package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Models
type FishProfile struct {
	ID              int    `json:"id" db:"id"`
	Name            string `json:"name" db:"name"`
	Variety         string `json:"variety" db:"variety"`
	Color           string `json:"color" db:"color"`
	Age             string `json:"age" db:"age"`
	HealthStatus    string `json:"healthStatus" db:"health_status"`
	LastHealthCheck string `json:"lastHealthCheck" db:"last_health_check"`
	Notes           string `json:"notes" db:"notes"`
	TankID          string `json:"tankId" db:"tank_id"`
}

type FeedingSchedule struct {
	ID       int    `json:"id"`
	Time     string `json:"time"`
	FoodType string `json:"foodType"`
	FishID   int    `json:"fishId"`
	FishName string `json:"fishName" db:"name"` // New field
}

// Database connection
var db *sql.DB

// Initialize database

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./fish_management.db")
	log.Println("Connecting to database...")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create tables
	createTables()
}

func createTables() {
	// Create fish profiles table
	log.Println("Creating database tables...")
	fishTable := `
		CREATE TABLE IF NOT EXISTS fish_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			variety TEXT NOT NULL,
			color TEXT NOT NULL,
			age TEXT NOT NULL,
			health_status TEXT NOT NULL,
			last_health_check TEXT NOT NULL,
			notes TEXT,
			tank_id TEXT NOT NULL
		);`

	// Create feeding schedules table
	feedingTable := `
		CREATE TABLE IF NOT EXISTS feeding_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT NOT NULL,
			food_type TEXT NOT NULL,
			fish_id TEXT NOT NULL,
			FOREIGN KEY (fish_id) REFERENCES fish_profiles(id) ON DELETE CASCADE
		);`

	if _, err := db.Exec(fishTable); err != nil {
		log.Fatal("Failed to create fish_profiles table:", err)
	}

	if _, err := db.Exec(feedingTable); err != nil {
		log.Fatal("Failed to create feeding_schedules table:", err)
	}

	log.Println("Database tables created successfully")
}

// Fish Profile Handlers

func createFish(c *gin.Context) {
	log.Println("create fish")

	var fish FishProfile
	if err := c.ShouldBindJSON(&fish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println(fish)

	query := `
			INSERT INTO fish_profiles (name, variety, color, age, health_status, last_health_check, notes, tank_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`

	_, err := db.Exec(query, fish.Name, fish.Variety, fish.Color, fish.Age,
		fish.HealthStatus, fish.LastHealthCheck, fish.Notes, fish.TankID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fish profile"})
		return
	}

	c.JSON(http.StatusCreated, fish)
}

func getFish(c *gin.Context) {
	id := c.Param("id")

	var fish FishProfile
	query := `
			SELECT id, name, variety, color, age, health_status, last_health_check, notes, tank_id
			FROM fish_profiles WHERE id = ?
		`

	err := db.QueryRow(query, id).Scan(
		&fish.ID, &fish.Name, &fish.Variety, &fish.Color, &fish.Age,
		&fish.HealthStatus, &fish.LastHealthCheck, &fish.Notes, &fish.TankID,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fish not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fish"})
		return
	}

	c.JSON(http.StatusOK, fish)
}

func getAllFish(c *gin.Context) {
	query := `
			SELECT id, name, variety, color, age, health_status, last_health_check, notes, tank_id
			FROM fish_profiles
		`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fish profiles"})
		return
	}
	defer rows.Close()

	var fishList []FishProfile
	for rows.Next() {
		var fish FishProfile
		err := rows.Scan(
			&fish.ID, &fish.Name, &fish.Variety, &fish.Color, &fish.Age,
			&fish.HealthStatus, &fish.LastHealthCheck, &fish.Notes, &fish.TankID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan fish data"})
			return
		}
		fishList = append(fishList, fish)
	}

	c.JSON(http.StatusOK, fishList)
}

func updateFish(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)

	var fish FishProfile
	if err := c.ShouldBindJSON(&fish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
			UPDATE fish_profiles
			SET name = ?, variety = ?, color = ?, age = ?, health_status = ?,
				last_health_check = ?, notes = ?, tank_id = ?
			WHERE id = ?
		`

	result, err := db.Exec(query, fish.Name, fish.Variety, fish.Color, fish.Age,
		fish.HealthStatus, fish.LastHealthCheck, fish.Notes, fish.TankID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fish profile"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fish not found"})
		return
	}

	fish.ID = id
	c.JSON(http.StatusOK, fish)
}

func deleteFish(c *gin.Context) {
	id := c.Param("id")

	query := `DELETE FROM fish_profiles WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete fish profile"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fish not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fish profile deleted successfully"})
}

// Feeding Schedule Handlers

func createFeedingSchedule(c *gin.Context) {
	var schedule FeedingSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `INSERT INTO feeding_schedules (time, food_type, fish_id) VALUES (?, ?, ?)`
	result, err := db.Exec(query, schedule.Time, schedule.FoodType, schedule.FishID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feeding schedule"})
		return
	}

	id, _ := result.LastInsertId()
	schedule.ID = int(id)

	c.JSON(http.StatusCreated, schedule)
}

func getFeedingSchedule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	var schedule FeedingSchedule
	query := `
	  SELECT 
		fs.id, 
		fs.time, 
		fs.food_type, 
		fs.fish_id,
		fp.name
	  FROM feeding_schedules fs
	  JOIN fish_profiles fp ON fs.fish_id = fp.id
	  WHERE fs.id = ?
	`

	err = db.QueryRow(query, id).Scan(&schedule.ID, &schedule.Time, &schedule.FoodType, &schedule.FishID, &schedule.FishName)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feeding schedule not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve feeding schedule"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func getAllFeedingSchedules(c *gin.Context) {
	query := `
	  SELECT 
		fs.id, 
		fs.time, 
		fs.food_type, 
		fs.fish_id,
		fp.name
	  FROM feeding_schedules fs
	  JOIN fish_profiles fp ON fs.fish_id = fp.id
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve feeding schedules"})
		return
	}
	defer rows.Close()

	var schedules []FeedingSchedule
	for rows.Next() {
		var schedule FeedingSchedule
		err := rows.Scan(&schedule.ID, &schedule.Time, &schedule.FoodType, &schedule.FishID, &schedule.FishName)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan schedule data"})
			return
		}
		schedules = append(schedules, schedule)
	}

	c.JSON(http.StatusOK, schedules)
}

func getFeedingSchedulesByFish(c *gin.Context) {
	fishID := c.Param("fishId")

	query := `SELECT id, time, food_type, fish_id FROM feeding_schedules WHERE fish_id = ?`

	rows, err := db.Query(query, fishID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve feeding schedules"})
		return
	}
	defer rows.Close()

	var schedules []FeedingSchedule
	for rows.Next() {
		var schedule FeedingSchedule
		err := rows.Scan(&schedule.ID, &schedule.Time, &schedule.FoodType, &schedule.FishID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan schedule data"})
			return
		}
		schedules = append(schedules, schedule)
	}

	c.JSON(http.StatusOK, schedules)
}

func updateFeedingSchedule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	var schedule FeedingSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE feeding_schedules SET time = ?, food_type = ?, fish_id = ? WHERE id = ?`
	result, err := db.Exec(query, schedule.Time, schedule.FoodType, schedule.FishID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feeding schedule"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feeding schedule not found"})
		return
	}

	schedule.ID = id
	c.JSON(http.StatusOK, schedule)
}

func deleteFeedingSchedule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	query := `DELETE FROM feeding_schedules WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feeding schedule"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feeding schedule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feeding schedule deleted successfully"})
}

// Health check endpoint

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"database":  "connected",
	})
}

// Setup routes

func setupRoutes() *gin.Engine {
	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Health check
		api.GET("/health", healthCheck)

		// Fish profile routes
		fish := api.Group("/fish")
		{
			fish.POST("", createFish)
			fish.GET("", getAllFish)
			fish.GET("/:id", getFish)
			fish.PUT("/:id", updateFish)
			fish.DELETE("/:id", deleteFish)
		}

		// Feeding schedule routes
		feeding := api.Group("/feeding-schedules")
		{
			feeding.POST("", createFeedingSchedule)
			feeding.GET("", getAllFeedingSchedules)
			feeding.GET("/:id", getFeedingSchedule)
			feeding.GET("/fish/:fishId", getFeedingSchedulesByFish)
			feeding.PUT("/:id", updateFeedingSchedule)
			feeding.DELETE("/:id", deleteFeedingSchedule)
		}
	}

	return r
}

func main() {
	// Initialize database
	log.Println("Initializing database...")
	initDB()
	defer db.Close()

	// Setup routes
	router := setupRoutes()

	// Start server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Printf("Health check available at: http://localhost%s/api/v1/health", port)
	log.Fatal(router.Run(port))
}
