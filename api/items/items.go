package items

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/IiMDMiI/MarketServer/pkg/dbservice"
)

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

var (
	mu sync.Mutex
)

func GetItems(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	items := fetchItemsFromDB(w)

	json.NewEncoder(w).Encode(items)
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Authorization")
	fmt.Println(key)
	if !isAdmin(key) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()

	_, err := dbservice.DB.Exec("INSERT INTO items (name, price) VALUES ($1, $2)", item.Name, item.Price)
	if err != nil {
		mu.Unlock()
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Item was created"))
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Authorization")
	if !isAdmin(key) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updatedItem Item
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	id := r.URL.Query().Get("id")
	result, err := dbservice.DB.Exec("UPDATE items SET name = $1, price = $2 WHERE item_id = $3", updatedItem.Name, updatedItem.Price, id)
	if err != nil {
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking updated rows", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Item was updated"))
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Authorization")
	if !isAdmin(key) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")

	mu.Lock()
	defer mu.Unlock()

	result, err := dbservice.DB.Exec("DELETE FROM items WHERE item_id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking deleted rows", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Item was deleted"))
}

func fetchItemsFromDB(w http.ResponseWriter) []Item {
	rows, err := dbservice.DB.Query("SELECT * FROM items")
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			http.Error(w, "Error scanning item", http.StatusInternalServerError)
			return nil
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over items", http.StatusInternalServerError)
		return nil
	}

	return items
}

func isAdmin(key string) bool {
	response := dbservice.DB.QueryRow("SELECT user_id FROM sessions WHERE token = $1", key)
	var userID int = -1

	err := response.Scan(&userID)
	if err != nil {
		return false
	}

	row := dbservice.DB.QueryRow("SELECT user_id FROM admins WHERE user_id = $1", userID)
	return row.Err() == nil
}
