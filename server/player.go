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
	playerGetIdByNameStmt *sql.Stmt
	playerDataStmts       = make(map[string]*sql.Stmt)
)

func preparePlayerStatements() error {
	sql := "select id from players where name = $1"
	var err error
	playerGetIdByNameStmt, err = db.Prepare(sql)
	if err != nil {
		return err
	}

	pointsQuery := "select extract(epoch from updated)::int, points from player_points_%s_%s where player_id = $1 order by updated"
	for _, kind := range [3]string{"points", "fleet", "research"} {
		for _, period := range [3]string{"week", "month", "all_time"} {
			stmt, err := db.Prepare(fmt.Sprintf(pointsQuery, kind, period))
			if err != nil {
				return err
			}
			playerDataStmts["points_"+kind+"_"+period] = stmt
		}
	}

	rankingQuery := "select extract(epoch from updated)::int, rank from player_ranking_%s_%s where player_id = $1 order by updated"
	for _, kind := range [3]string{"points", "fleet", "research"} {
		for _, period := range [3]string{"week", "month", "all_time"} {
			stmt, err := db.Prepare(fmt.Sprintf(rankingQuery, kind, period))
			if err != nil {
				return err
			}
			playerDataStmts["ranking_"+kind+"_"+period] = stmt
		}
	}

	return nil
}

func playerCtrl(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")

	row := playerGetIdByNameStmt.QueryRow(name)
	var playerId int
	if err := row.Scan(&playerId); err != nil {
		w.Write(easterEgg61)
		return
	}

	data := make(map[string]string)
	data["Controller"] = "player"
	data["Name"] = name

	for _, what := range [2]string{"points", "ranking"} {
		for _, kind := range [3]string{"points", "fleet", "research"} {
			for _, period := range [3]string{"week", "month", "all_time"} {
				key := what + "_" + kind + "_" + period
				stmt := playerDataStmts[key]
				rows, err := stmt.Query(playerId)
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

	if err := playerTmpl.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}
