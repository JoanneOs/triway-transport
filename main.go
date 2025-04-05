package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/gorilla/mux"
)

// Database models
type Truck struct {
	ID     uint   `gorm:"primaryKey"`
	Make   string `gorm:"size:50"`
	Model  string `gorm:"size:50"`
	VIN    string `gorm:"size:17;unique"`
	Status string `gorm:"size:20"` // Active, Maintenance
}

type LocationHistory struct {
	ID        uint      `gorm:"primaryKey"`
	TruckID   uint      `gorm:"index"`
	Latitude  float64   `gorm:"type:decimal(10,8)"`
	Longitude float64   `gorm:"type:decimal(11,8)"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

var db *gorm.DB

func main() {
	// 1. Connect to MySQL
	dsn := "root:4101984jojoTT!@tcp(127.0.0.1:3306)/triway_transport?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	fmt.Println("Connected to database!")

	// 2. Create tables
	db.AutoMigrate(&Truck{}, &LocationHistory{})

	// 3. Set up routes
	r := mux.NewRouter()
	
	// Truck endpoints
	r.HandleFunc("/api/trucks", getTrucks).Methods("GET")
	r.HandleFunc("/api/trucks", createTruck).Methods("POST")
	r.HandleFunc("/api/trucks/{id}", updateTruckStatus).Methods("PUT")
	
	// Location endpoints
	r.HandleFunc("/api/locations", recordLocation).Methods("POST")
	r.HandleFunc("/api/locations/{truckId}", getLocationHistory).Methods("GET")
	
	// Serve frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	
	// 4. Start server
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handler functions
func getTrucks(w http.ResponseWriter, r *http.Request) {
	var trucks []Truck
	db.Find(&trucks)
	json.NewEncoder(w).Encode(trucks)
}

func createTruck(w http.ResponseWriter, r *http.Request) {
	var truck Truck
	if err := json.NewDecoder(r.Body).Decode(&truck); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if result := db.Create(&truck); result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(truck)
}

func updateTruckStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var update struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	db.Model(&Truck{}).Where("id = ?", id).Update("status", update.Status)
	w.Write([]byte("Truck status updated"))
}

func recordLocation(w http.ResponseWriter, r *http.Request) {
	var loc LocationHistory
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	db.Create(&loc)
	w.Write([]byte("Location recorded"))
}

func getLocationHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	truckId := vars["truckId"]
	
	var history []LocationHistory
	db.Where("truck_id = ?", truckId).Find(&history)
	json.NewEncoder(w).Encode(history)
}