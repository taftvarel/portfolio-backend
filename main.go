package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var db *sql.DB

// Project represents a portfolio project
type Project struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tech        []string `json:"tech"`
	GitHub      string   `json:"github"`
	Demo        string   `json:"demo"`
	ImageURL    string   `json:"image_url,omitempty"`
}

// Profile represents user profile information
type Profile struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	Bio      string   `json:"bio"`
	Email    string   `json:"email"`
	GitHub   string   `json:"github"`
	LinkedIn string   `json:"linkedin"`
	Skills   []string `json:"skills"`
	ImageURL string   `json:"image_url,omitempty"`
}

// ContactMessage represents a contact form submission
type ContactMessage struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// --- Handlers (you can later replace in-memory with MySQL queries) ---

var profile = Profile{
	Name:     "John Doe",
	Title:    "Full Stack Developer",
	Bio:      "Passionate about creating innovative web applications with modern technologies. Specialized in React.js frontend development and Go backend services.",
	Email:    "john.doe@example.com",
	GitHub:   "https://github.com/johndoe",
	LinkedIn: "https://linkedin.com/in/john-doe",
	Skills:   []string{"Go", "React", "JavaScript", "TypeScript", "PostgreSQL", "MongoDB", "Docker", "AWS"},
	ImageURL: "https://via.placeholder.com/150",
}

// --- Fetch Profile from MySQL ---
func getProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var p Profile
	var skills []string

	// Fetch profile (assuming only one profile with id=1)
	row := db.QueryRow("SELECT name, title, bio, email, github, linkedin, image_url FROM profile WHERE id = 1")
	if err := row.Scan(&p.Name, &p.Title, &p.Bio, &p.Email, &p.GitHub, &p.LinkedIn, &p.ImageURL); err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Fetch skills
	rows, err := db.Query("SELECT skill FROM skills WHERE profile_id = 1")
	if err != nil {
		http.Error(w, "Error fetching skills", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err == nil {
			skills = append(skills, skill)
		}
	}
	p.Skills = skills

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    p,
	})
}

// --- Fetch All Projects from MySQL ---
func getProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT id, title, description, github, demo, image_url FROM projects")
	if err != nil {
		http.Error(w, "Error fetching projects", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var pr Project
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.Description, &pr.GitHub, &pr.Demo, &pr.ImageURL); err == nil {
			// Fetch tech stack for this project
			techRows, _ := db.Query("SELECT tech FROM project_tech WHERE project_id = ?", pr.ID)
			var techs []string
			for techRows.Next() {
				var tech string
				techRows.Scan(&tech)
				techs = append(techs, tech)
			}
			techRows.Close()
			pr.Tech = techs

			projects = append(projects, pr)
		}
	}

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Projects retrieved successfully",
		Data:    projects,
	})
}

// --- Fetch Single Project by ID ---
func getProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var pr Project
	row := db.QueryRow("SELECT id, title, description, github, demo, image_url FROM projects WHERE id = ?", id)
	if err := row.Scan(&pr.ID, &pr.Title, &pr.Description, &pr.GitHub, &pr.Demo, &pr.ImageURL); err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Fetch tech stack
	techRows, _ := db.Query("SELECT tech FROM project_tech WHERE project_id = ?", id)
	var techs []string
	for techRows.Next() {
		var tech string
		techRows.Scan(&tech)
		techs = append(techs, tech)
	}
	techRows.Close()
	pr.Tech = techs

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Project retrieved successfully",
		Data:    pr,
	})
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	var msg ContactMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Contact form submission: %+v", msg)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "Message sent successfully"})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Success: true, Message: "API is running", Data: map[string]string{"status": "healthy"}})
}

// --- main ---
func main() {
	// ✅ Setup MySQL connection inside main
	dsn := "root:varel0306@tcp(my-app-db-rds.cu9squsao8nh.us-east-1.rds.amazonaws.com:3306)/portfolio_db"
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("MySQL unreachable:", err)
	}
	log.Println("Connected to MySQL!")

	// Setup router
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/profile", getProfile).Methods("GET")
	api.HandleFunc("/projects", getProjects).Methods("GET")
	api.HandleFunc("/projects/{id:[0-9]+}", getProject).Methods("GET")
	api.HandleFunc("/contact", handleContact).Methods("POST")
	api.HandleFunc("/health", healthCheck).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://www.propcloud.fun"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", c.Handler(r)); err != nil {
		log.Fatal(err)
	}
}
