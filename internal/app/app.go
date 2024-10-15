package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/IiMDMiI/MarketServer/api/auth"
	"github.com/IiMDMiI/MarketServer/api/items"
	"github.com/IiMDMiI/MarketServer/pkg/dbservice"

	_ "github.com/lib/pq"
)

const DB_SETTINGS_PATH = "configs/dbsettings.yaml"

type App struct {
}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	bindRoutes()

	db := openDB()
	dbservice.DB = db

	defer db.Close()

	port := ":8080"
	println("Server is running on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
	return nil
}

func bindRoutes() {
	prefix := "/api/v1"
	//authenticating
	http.HandleFunc("POST "+prefix+"/login", auth.Login)
	http.HandleFunc("POST "+prefix+"/register", auth.Register)

	//items
	http.HandleFunc("POST "+prefix+"/item", items.CreateItem)
	http.HandleFunc("GET "+prefix+"/items", items.GetItems)
	http.HandleFunc("PUT "+prefix+"/item", items.UpdateItem)
	http.HandleFunc("DELETE "+prefix+"/item", items.DeleteItem)
}

func openDB() *dbservice.DBConnectorImpl {
	s := dbservice.New(DB_SETTINGS_PATH)
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbservice.Host(s), dbservice.Port(s), dbservice.User(s),
		dbservice.Password(s), dbservice.DBname(s))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return &dbservice.DBConnectorImpl{DB: db}
}
