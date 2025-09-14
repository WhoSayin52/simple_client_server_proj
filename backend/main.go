package main

import (
	"fmt"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("../frontend/"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Serving:", r.URL.Path)
		fs.ServeHTTP(w, r)
	})

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)

	fmt.Println("Server running at http://localhost:8080/n")
	http.ListenAndServe(":8080", nil)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	fmt.Printf("Received username: %s/n", username)
	password := r.FormValue("password")
	fmt.Printf("Received password: %s/n", password)

	w.Write([]byte("Username: " + username + "\n"))
	w.Write([]byte("Password: " + password))
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	fmt.Printf("Received username: %s/n", username)
	password := r.FormValue("password")
	fmt.Printf("Received password: %s/n", password)
	confirmPassword := r.FormValue("confirm-password")
	fmt.Printf("Received confirmed-password: %s/n", confirmPassword)

	if password != confirmPassword {
		w.Write([]byte("Passwords do not match"))
		return
	} else {
		w.Write([]byte("Welcome " + username))
	}
}
