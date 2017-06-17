package main

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"regexp"
)

var (
	typeaheadPlayerStmt   *sqlx.Stmt
	typeaheadAllianceStmt *sqlx.Stmt
	isVaildQuery          = regexp.MustCompile(`^[a-zA-Z0-9 ]+$`).MatchString
)

func prepareTypeaheadStatements() error {
	var err error

	playerSql := "select name from players where lower(name) like $1 limit 10"
	typeaheadPlayerStmt, err = db.Preparex(playerSql)
	if err != nil {
		return err
	}

	allianceSql := "select tag, name from alliances where lower(tag) like $1 or lower(name) like $1 limit 10"
	typeaheadAllianceStmt, err = db.Preparex(allianceSql)
	if err != nil {
		return err
	}

	return nil
}

type typeaheadPlayerRow struct {
	Name string `json:"name"`
}

type typeaheadAllianceRow struct {
	Tag  string `json:"tag"`
	Name string `json:"name"`
}

func typeaheadCtrl(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query().Get("q")
	if !isVaildQuery(q) {
		w.Write([]byte("{}"))
		return
	}
	param := q + "%"

	data := make(map[string]interface{})

	playerRows := []typeaheadPlayerRow{}
	if err := typeaheadPlayerStmt.Select(&playerRows, param); err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	data["players"] = playerRows

	allianceRows := []typeaheadAllianceRow{}
	if err := typeaheadAllianceStmt.Select(&allianceRows, param); err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	data["alliances"] = allianceRows

	encoded, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(encoded)
}
