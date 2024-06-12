package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var users []User

func init() {
	file, err := os.Open("csvファイルパス")
	if err != nil {
		log.Fatalf("Failed to open spreadsheet file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read spreadsheet data: %v", err)
	}

	for _, record := range records {
		if len(record) >= 2 {
			user := User{
				Email:    record[0],
				Password: record[1],
			}
			users = append(users, user)
		}
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, user := range users {
		if user.Email == newUser.Email {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
	}

	file, err := os.OpenFile("csvファイルパス", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{newUser.Email, newUser.Password})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	users = append(users, newUser)
	w.WriteHeader(http.StatusOK)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginUser User
	err := json.NewDecoder(r.Body).Decode(&loginUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, user := range users {
		if user.Email == loginUser.Email && user.Password == loginUser.Password {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Invalid email or password", http.StatusUnauthorized)
}

func main() {
	corsHandler := cors.Default().Handler(http.DefaultServeMux)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
