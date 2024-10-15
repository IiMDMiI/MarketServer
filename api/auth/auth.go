package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/IiMDMiI/MarketServer/pkg/dbservice"
	"github.com/IiMDMiI/MarketServer/pkg/tockengen"

	"golang.org/x/crypto/bcrypt"
)

var (
	mu           sync.Mutex
	passProvider PasswordProvider
	jwtSecret    []byte
)

func init() {
	passProvider = &DBPasswordProvider{}

	var err error
	jwtSecret, err = os.ReadFile("configs/jwtSecret.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var u User
	if !isAbleToRegister(r, w, &u) {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	_, err = dbservice.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", u.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User was registered")
}

func Login(w http.ResponseWriter, r *http.Request) {
	var u User
	if !isValidUser(r, w, &u) {
		return
	}

	password, err := passProvider.GetPassword(u.Username)
	if err != nil || passProvider.ValidateHash([]byte(u.Password), password) != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	tocken, err := tockengen.CreateToken(u.Username, jwtSecret)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User was logged in with token: %v", tocken)
	saveJwt(w, tocken, u.Username)
}
func isAbleToRegister(r *http.Request, w http.ResponseWriter, u *User) bool {
	if !isValidUser(r, w, u) {
		return false
	}

	if _, err := passProvider.GetPassword(u.Username); err == nil {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return false
	}
	return true
}

func isValidUser(r *http.Request, w http.ResponseWriter, u *User) bool {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if u.Username == "" || u.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return false
	}
	return true
}

type PasswordProvider interface {
	GetPassword(username string) ([]byte, error)
	ValidateHash(password, hash []byte) error
}

type DBPasswordProvider struct{}

func (p *DBPasswordProvider) GetPassword(username string) ([]byte, error) {
	var password string
	query := "SELECT password FROM users WHERE username = $1"
	err := dbservice.DB.QueryRow(query, username).Scan(&password)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no user found with username: %s", username)
	} else if err != nil {
		return nil, err
	}

	return []byte(password), nil
}
func (p *DBPasswordProvider) ValidateHash(password, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}

func saveJwt(w http.ResponseWriter, token, name string) {
	var userID int

	err := dbservice.DB.QueryRow("SELECT user_id FROM users WHERE username = $1", name).Scan(&userID)
	if err != nil {
		http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
		return
	}

	_, err = dbservice.DB.Exec("INSERT INTO sessions (user_id, token) VALUES ($1, $2)", userID, token)
	if err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}
}
