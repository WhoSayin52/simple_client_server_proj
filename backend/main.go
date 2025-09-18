package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	database *pgxpool.Pool
}

func main() {

	pool, err := initDB()
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	defer pool.Close()

	var s *Server = &Server{
		database: pool,
	}

	log.Println("Connection to database successful")

	fs := http.FileServer(http.Dir("../frontend/"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving:", r.URL.Path)
		fs.ServeHTTP(w, r)
	})

	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/signup", s.signupHandler)

	log.Println("Server running at http://localhost:8080/")
	http.ListenAndServe(":8080", nil)

}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	log.Printf("Received username: %s\n", username)
	password := r.FormValue("password")
	log.Printf("Received password: %s\n", password)

	ctx := r.Context()

	var storedPassword string
	err := s.database.QueryRow(ctx, "SELECT password FROM users WHERE username=$1", username).Scan(&storedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Username not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("DB error fetching password: %v\n", err)
		return
	}

	if password != storedPassword {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Login successful for user: " + username))
	w.Write([]byte("\nPassword " + password))
}

func (s *Server) signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	log.Printf("Received username: %s\n", username)
	password := r.FormValue("password")
	log.Printf("Received password: %s\n", password)
	confirmPassword := r.FormValue("confirm-password")
	log.Printf("Received confirmed-password: %s\n", confirmPassword)

	if password != confirmPassword {
		w.Write([]byte("Passwords do not match"))
		return
	}

	ctx := r.Context()

	var exists bool
	err := s.database.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("DB error checking username existence: %v\n", err)
		return
	}

	if exists {
		w.Write([]byte("User already exists"))
		return
	}

	_, err = s.database.Exec(ctx, "INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		log.Printf("DB error inserting user: %v\n", err)
		return
	}

	w.Write([]byte("Signup successful " + username))
	w.Write([]byte("\nPassword " + password))
}

func initDB() (*pgxpool.Pool, error) {
	ctx := context.Background()

	log.Println("Connecting to postgres database...")
	adminConn, err := pgx.Connect(ctx, "postgres://test_db:test_db@localhost:5432/postgres")
	if err != nil {
		return nil, fmt.Errorf("unable to connect database: %w", err)
	}
	defer adminConn.Close(ctx)

	dbName := "test_db"

	var exists bool
	err = adminConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbName).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	log.Printf("Connecting to %s database...\n", dbName)
	if !exists {
		_, err = adminConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	dbPool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://test_db:test_db@localhost:5432/%s", dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}

	log.Printf("Database running at postgres://test_db:***@localhost:5432/%s\n", dbName)

	return dbPool, nil
}
