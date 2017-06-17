package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

var (
	allianceExistsStmt *sql.Stmt
	allianceDataStmts  = make(map[string]*sql.Stmt)
)

func prepareAllianceStatements() error {
	sql := "select 1 from alliances where tag = $1"
	var err error
	allianceExistsStmt, err = db.Prepare(sql)
	if err != nil {
		return err
	}

	pointsQuery := "select extract(epoch from updated)::int, points from alliance_points_%s_%s where alliance_tag = $1 order by updated"
	for _, kind := range [3]string{"points", "fleet", "research"} {
		for _, period := range [3]string{"week", "month", "all_time"} {
			stmt, err := db.Prepare(fmt.Sprintf(pointsQuery, kind, period))
			if err != nil {
				return err
			}
			allianceDataStmts["points_"+kind+"_"+period] = stmt
		}
	}

	rankingQuery := "select extract(epoch from updated)::int, rank from alliance_ranking_%s_%s where alliance_tag = $1 order by updated"
	for _, kind := range [3]string{"points", "fleet", "research"} {
		for _, period := range [3]string{"week", "month", "all_time"} {
			stmt, err := db.Prepare(fmt.Sprintf(rankingQuery, kind, period))
			if err != nil {
				return err
			}
			allianceDataStmts["ranking_"+kind+"_"+period] = stmt
		}
	}

	return nil
}

func allianceCtrl(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tag := ps.ByName("tag")

	row := allianceExistsStmt.QueryRow(tag)
	var dummy int
	if err := row.Scan(&dummy); err != nil {
		w.Write(easterEgg62)
		return
	}

	data := make(map[string]string)
	data["Controller"] = "alliance"
	data["Tag"] = tag

	for _, what := range [2]string{"points", "ranking"} {
		for _, kind := range [3]string{"points", "fleet", "research"} {
			for _, period := range [3]string{"week", "month", "all_time"} {
				key := what + "_" + kind + "_" + period
				stmt := allianceDataStmts[key]
				rows, err := stmt.Query(tag)
				if err != nil {
					log.Println(err)
					http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
					return
				}
				pairs := [][2]int{}
				for rows.Next() {
					var t int
					var n int
					if err := rows.Scan(&t, &n); err != nil {
						log.Println(err)
						http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
						return
					}
					pairs = append(pairs, [2]int{t, n})
				}
				if err := rows.Close(); err != nil {
					log.Println(err)
					http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
					return
				}
				encoded, err := json.Marshal(pairs)
				if err != nil {
					log.Println(err)
					http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
					return
				}
				data[key] = string(encoded)
			}
		}
	}

	if err := allianceTmpl.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}
