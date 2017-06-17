package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type topRow struct {
	Name                string  `json:"name"`
	AllianceTag         *string `db:"alliance_tag" json:"alliance_tag"`
	Points              int     `json:"points"`
	Rank                int     `json:"rank"`
	WeekDifference      *int    `db:"week_difference" json:"week_difference"`
	WeekDifferenceRank  *int    `db:"week_difference_rank" json:"week_difference_rank"`
	MonthDifference     *int    `db:"month_difference" json:"month_difference"`
	MonthDifferenceRank *int    `db:"month_difference_rank" json:"month_difference_rank"`
}

func topCtrl(kind string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		query := r.URL.Query()

		var orderBy string
		switch query.Get("order-by") {
		case "":
			orderBy = "points"
		case "points":
			orderBy = "points"
		case "week-difference":
			orderBy = "week_difference"
		case "month-difference":
			orderBy = "month_difference"
		default:
			w.Write(easterEgg41)
			return
		}

		var direction string
		switch query.Get("direction") {
		case "":
			direction = "desc"
		case "desc":
			direction = "desc"
		case "asc":
			direction = "asc"
		default:
			w.Write(easterEgg42)
			return
		}

		sql := "select count(1) from top_" + kind
		var count int
		if err := db.QueryRow(sql).Scan(&count); err != nil {
			log.Println(err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		numPages := (count + 99) / 100

		var page int
		pageStr := query.Get("page")
		if pageStr == "" {
			page = 1
		} else {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				w.Write(easterEgg43)
				return
			} else if page < 1 || page > numPages {
				w.Write(easterEgg44)
				return
			}
		}
		offset := (page - 1) * 100

		rows := []topRow{}
		sql = "select * from top_%s order by %s %s nulls last limit 100 offset %d"
		sql = fmt.Sprintf(sql, kind, orderBy, direction, offset)
		if err := db.Select(&rows, sql); err != nil {
			log.Println(err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := make(map[string]interface{})
		data["Controller"] = "top" + kind
		data["OrderBy"] = strings.Replace(orderBy, "_", "-", 1)
		data["Direction"] = direction
		data["NumPages"] = numPages
		data["Page"] = page
		data["Rows"] = rows
		if err := topTmpl.Execute(w, data); err != nil {
			log.Println(err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func apiTopCtrl(kind string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		rows := []topRow{}
		sql := "select * from top_" + kind
		if err := db.Select(&rows, sql); err != nil {
			log.Println(err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		encoded, err := json.Marshal(rows)
		if err != nil {
			log.Println(err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(encoded)
	}
}
