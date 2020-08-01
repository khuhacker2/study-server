package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)

var database *dbr.Connection

func main() {
	db, err := dbr.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", configs.Database.User, configs.Database.Password, configs.Database.Host, configs.Database.Name), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	database = db
	defer database.Close()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/users/:no", GetUsers),
		rest.Post("/users", PostUsers),
		rest.Post("/token", PostToken),
		rest.Get("/studygroups/:no", GetStudygroup),
		rest.Post("/studygroups", PostStudygroup),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configs.Port), api.MakeHandler()))
}
