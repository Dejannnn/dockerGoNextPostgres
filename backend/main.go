package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/helth", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("eee"))
	}).Methods("GET")

	router.HandleFunc("/api/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/api/users/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/api/users", createUsers(db)).Methods("POST")
	router.HandleFunc("/api/users/{id}", updateUsers(db)).Methods("PUT")
	router.HandleFunc("/api/users", deleteUsers(db)).Methods("DELETE")

	enahnceRouter := enableCors(jsonContentTypeMiddleware(router))
	log.Fatal(http.ListenAndServe(":3000", enahnceRouter))
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Auhtorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appliaction/json")
		next.ServeHTTP(w, r)
	})
}

func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			http.Error(w, "Faild to fetch users", http.StatusInternalServerError)
			log.Println("Error fetching useers: ", err)
			return
		}
		defer rows.Close()
		var users []User

		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
				http.Error(w, "Faild to scan user", http.StatusInternalServerError)
				log.Println("Error scanning", err)
				return
			}
			users = append(users, user)
		}
		if err := json.NewEncoder(w).Encode(users); err != nil {
			http.Error(w, "Failed to encode users", http.StatusInternalServerError)
			log.Println("Error encoding users:", err)
			return
		}
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		user_id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Error while parsing parametar usrr_id", http.StatusInternalServerError)
			return
		}
		var user User
		err = db.QueryRow("SELECT * FROM users WHERE id = $1", user_id).Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			http.Error(w, "Faild to fetch user", http.StatusNotFound)
			log.Println("Error fetching useers: ", err)
			return
		}

		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "Failed to encode users", http.StatusInternalServerError)
			log.Println("Error encoding users:", err)
			return
		}
	}
}

func createUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var user User
		if err := decoder.Decode(&user); err != nil {
			http.Error(w, "Error while decode body", http.StatusInternalServerError)
			return
		}

		err := db.QueryRow("INSERT INTO users(name, email) VALUES($1,$2) RETURNING ID", user.Name, user.Email).Scan(&user.ID)
		if err != nil {
			http.Error(w, "Creat user faild", http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "Failed to encode users", http.StatusInternalServerError)
			log.Println("Error encoding users:", err)
			return
		}
	}
}

func updateUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user_id := vars["id"]
		var user User

		// Execute the update query
		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", user.Name, user.Email, user_id)
		if err != nil {
			log.Fatal(err)
		}

		// Retrieve the updated user data from the database
		var updatedUser User
		err = db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", user_id).Scan(&updatedUser.ID, &updatedUser.Name, &updatedUser.Email)
		if err != nil {
			log.Fatal(err)
		}

		// Send the updated user data in the response
		json.NewEncoder(w).Encode(updatedUser)
	}
}
func deleteUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		err := db.QueryRow("DELETE * FROM users WHERE id= $1", id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode("Deleted user")
	}
}
