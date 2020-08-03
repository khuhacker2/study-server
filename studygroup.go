package main

import (
	"strconv"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

type Studygroup struct {
	No          uint64    `json:"no" db:"no"`
	Category    int       `json:"category" db:"category"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func (study *Studygroup) Get() {
	database.NewSession(nil).Select("*").From("studygroups").Where("no=?", study.No).Load(study)
}

func GetStudygroup(w rest.ResponseWriter, r *rest.Request) {
	no, _ := strconv.ParseUint(r.PathParam("no"), 10, 64)
	study := Studygroup{No: no}
	study.Get()

	w.WriteJson(study)
}

func PostStudygroup(w rest.ResponseWriter, r *rest.Request) {
	authHeader := r.Header["Authorization"]
	if authHeader == nil || len(authHeader) == 0 || len(authHeader[0]) < len("Bearer ") {
		writeAuthError(w)
		return
	}

	no, ok := parseToken(authHeader[0][len("Bearer "):])
	if !ok {
		writeAuthError(w)
		return
	}

	props := map[string]interface{}{}
	r.DecodeJsonPayload(&props)

	tr, _ := database.NewSession(nil).Begin()
	defer tr.RollbackUnlessCommitted()

	res, err := tr.InsertInto("studygroups").
		Columns("category", "name", "description").
		Values(props["category"], props["name"], props["description"]).Exec()

	if err != nil {
		return
	}

	groupNo, _ := res.LastInsertId()
	_, err = tr.InsertInto("study_members").Columns("studygroup", "user").Values(groupNo, no).Exec()
	if err != nil {
		return
	}

	tr.Commit()

	study := Studygroup{No: uint64(groupNo)}
	study.Get()
	w.WriteJson(study)
}

type Article struct {
	No         uint64    `json:"no" db:"no"`
	Studygroup uint64    `json:"studygroup" db:"studygroup"`
	Author     uint64    `json:"author" db:"author"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (article *Article) Get() {
	database.NewSession(nil).Select("*").From("articles").Where("no=?", article.No).Load(article)
}

func GetArticle(w rest.ResponseWriter, r *rest.Request) {
	no, _ := strconv.ParseUint(r.PathParam("no"), 10, 64)

	article := Article{No: no}
	article.Get()
	w.WriteJson(article)
}

func PostArticle(w rest.ResponseWriter, r *rest.Request) {
	authHeader := r.Header["Authorization"]
	if authHeader == nil || len(authHeader) == 0 || len(authHeader[0]) < len("Bearer ") {
		writeAuthError(w)
		return
	}

	no, ok := parseToken(authHeader[0][len("Bearer "):])
	if !ok {
		writeAuthError(w)
		return
	}

	props := map[string]interface{}{}
	r.DecodeJsonPayload(&props)

	joined := 0
	session := database.NewSession(nil)
	session.Select("1").From("study_members").Where("user=? AND studygroup=?", no, props["studygroup"]).Load(&joined)
	if joined != 1 {
		return
	}

	res, err := session.InsertInto("articles").Columns("studygroup", "author", "title", "content").Values(props["studygroup"], no, props["title"], props["content"]).Exec()
	if err != nil {
		return
	}

	articleNo, _ := res.LastInsertId()

	article := Article{No: uint64(articleNo)}
	article.Get()
	w.WriteJson(article)
}
