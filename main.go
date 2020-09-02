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
	router, err := rest.MakeRouter(
		rest.Get("/users/:no", GetUsers),
		rest.Post("/users", PostUsers),
		rest.Post("/token", PostToken),
		rest.Get("/studygroups/rank", GetStudygroupRank),
		rest.Get("/studygroups/:no", GetStudygroup),
		rest.Post("/studygroups", PostStudygroup),
		rest.Get("/studygroups/:no/articles", GetStudygroupsArticles),
		rest.Get("/me", GetMe),
		rest.Put("/me", PutMe),
		rest.Get("/me/studygroups", GetMeStudygroups),
		rest.Get("/articles/:no", GetArticle),
		rest.Post("/articles", PostArticle),
		rest.Get("/articles/:no/comments", GetArticlesComments),
		rest.Get("/comments/:no", GetComment),
		rest.Post("/comments", PostComment),
		rest.Post("/images", PostImage),
		rest.Get("/ws", GetWS),
	)

	if err != nil {
		log.Fatal(err)
	}

	api.SetApp(router)

	http.Handle("/", api.MakeHandler())
	http.Handle("/files/", http.StripPrefix("/files", http.FileServer(http.Dir("files"))))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configs.Port), nil))
}
