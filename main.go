package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	ID       int    `json:"ID"`
	Name     string `json:"Name"`
	Age      int    `json:"Age"`
	Gender   string `json:"Gender"`
	Email    string `json:"Email"`
	Cart     int    `json:"Cart"`
	Role     string `json:"Role"`
	Password string `json:"Password"`
}

func checkEmailUser(user User) (int, error) {
	if len(user.Email) <= 10 {
		return 0, nil
	}
	check_Email, err := db.Query("SELECT Email FROM users WHERE Email = $1", user.Email)
	if err != nil {
		return 0, err
	}
	defer check_Email.Close()
	var check string
	if check_Email.Next() {
		err = check_Email.Scan(&check)
		if err != nil {
			return 0, err
		}
	}
	if check != "" {
		return 1, nil
	}
	return 2, nil
}

func checkPassWord(user User) int {
	if len(user.Password) < 8 {
		return 0
	}
	check := []int{0, 0, 0}
	for _, k := range user.Password {
		if k >= 48 && k <= 57 {
			check[0]++
		} else if k >= 65 && k <= 90 || k >= 97 && k <= 122 {
			check[1]++
		} else if k >= 33 && k <= 126 {
			check[2]++
		}
	}
	for _, k := range check {
		if k == 0 {
			return 0
		}
	}
	return 1
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	resultCheckEmail, err := checkEmailUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resultCheckEmail == 0 {
		http.Error(w, "Email phải lớn hơn 10 ký tự", http.StatusBadRequest)
		return
	} else if resultCheckEmail == 1 {
		http.Error(w, "Email đã tồn tại", http.StatusBadRequest)
		return
	}

	resultCheckPassWord := checkPassWord(user)

	if resultCheckPassWord == 0 {
		http.Error(w, "Password phải có hơn 8 ký tự với ít nhất một số , một từ và một ký tự đặc biệt", http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("INSERT INTO users (name, age, gender, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id", user.Name, user.Age, user.Gender, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func main() {
	connStr := "user=postgres password=admin123 dbname=Food sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := mux.NewRouter()
	r.HandleFunc("/users", createUser).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}