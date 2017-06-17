package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/GeertJohan/go.rice"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type logConfig struct {
	Filename string `toml:"filename"`
}

type databaseConfig struct {
	DataSource string `toml:"data_source"`
}

type serverConfig struct {
	Host string
	Port int
}

type config struct {
	Log      logConfig
	Database databaseConfig
	Server   serverConfig
}

var (
	db             *sqlx.DB
	configFilename = flag.String("config", "", "Config filename")
	topTmpl        *template.Template
	allianceTmpl   *template.Template
	playerTmpl     *template.Template
)

func initTemplates() {
	wd, _ := os.Getwd()

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	// Why use first given filename as name? Go there:
	// https://stackoverflow.com/questions/10199219/go-template-function
	topTmpl = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles(
		filepath.Join(wd, "./templates/base.html"),
		filepath.Join(wd, "./templates/top.html")))
	allianceTmpl = template.Must(template.ParseFiles(
		filepath.Join(wd, "./templates/base.html"),
		filepath.Join(wd, "./templates/alliance.html")))
	playerTmpl = template.Must(template.ParseFiles(
		filepath.Join(wd, "./templates/base.html"),
		filepath.Join(wd, "./templates/player.html")))
}

func addSecurityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Referrer-Policy", "no-referrer")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-Frame-Options", "DENY")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
		h.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()
	if *configFilename == "" {
		log.Fatal("missing config flag")
	}

	var config config
	if _, err := toml.DecodeFile(*configFilename, &config); err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(config.Log.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	log.Println("initializing templates")
	initTemplates()

	log.Println("connecting to database")
	db, err = sqlx.Connect("postgres", config.Database.DataSource)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("preparing statements")
	if err := preparePlayerStatements(); err != nil {
		log.Fatal(err)
	}
	if err := prepareAllianceStatements(); err != nil {
		log.Fatal(err)
	}
	if err := prepareTypeaheadStatements(); err != nil {
		log.Fatal(err)
	}

	log.Println("setting http router up")
	router := httprouter.New()
	router.GET("/", topCtrl("points"))
	router.GET("/top-fleets", topCtrl("fleet"))
	router.GET("/alliance/:tag", allianceCtrl)
	router.GET("/player/:name", playerCtrl)
	router.GET("/typeahead", typeaheadCtrl)
	router.GET("/api/top-points", apiTopCtrl("points"))
	router.GET("/api/top-fleets", apiTopCtrl("fleet"))
	http.Handle("/", addSecurityHeaders(router))

	log.Println("setting static file router up")
	box := rice.MustFindBox("static")
	staticServer := http.StripPrefix("/static/", http.FileServer(box.HTTPBox()))
	http.Handle("/static/", staticServer)

	addr := config.Server.Host + ":" + strconv.Itoa(config.Server.Port)
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
