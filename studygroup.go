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
