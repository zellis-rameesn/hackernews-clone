package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/golangcollege/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type Application struct {
	errorLog    *log.Logger
	infoLog     *log.Logger
	userRepo    *UserRepo
	postRepo    *PostRepo
	templateDir string
	publicDir   string
	tempRender  *TemplateRenderer
	session     *sessions.Session
}

var session *sessions.Session
var secret = []byte("u46IpCV9y5Vlur8YvODJEhgOY8m9JAS2")

func main() {
	dbName := "app.db"

	db, err := connectToDb(dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	session = sessions.New(secret)
	session.Lifetime = 24 * time.Hour

	app := &Application{
		errorLog:    log.New(os.Stderr, "ERROR\t", log.Ltime|log.LstdFlags|log.Lmicroseconds|log.Lshortfile),
		infoLog:     log.New(os.Stdout, "INFO\t", log.Ltime|log.LstdFlags),
		userRepo:    NewUserRepo(db),
		postRepo:    NewPostRepo(db),
		templateDir: "./templates",
		publicDir:   "./public",
		session:     session,
	}
	newTempRenderer := NewTemplateRenderer(app.templateDir, true)
	app.tempRender = newTempRenderer
	app.Serve()
}

func connectToDb(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	return db, nil
}
