package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/sirupsen/logrus"
)

// dbName is the SQLite database path, overridable with DB_PATH.
var dbName = envOr("DB_PATH", "master.db")

// App is App
type App struct {
	conf    *Config
	srv     *http.Server
	l       *logrus.Logger
	dbLayer *dbLayer
}

type dbLayer struct {
	sync.Mutex
	db *sql.DB
}

func main() {
	l := initLogger()
	l.Infoln("hi")

	l.Infoln("reading config...")
	conf := loadConfig(l)
	l.Infoln("config loaded")

	l.Infoln("connecting to the database...")
	db, err := openDatabase(l)
	panicErr(l, err)
	l.Infoln("database connection estabilished")

	l.Infoln("initializing the server")
	router := mux.NewRouter()
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", envOr("BIND_ADDR", "127.0.0.1"), conf.Port),
		Handler: logRequestHandler(l, router),
	}

	a := &App{
		conf:    conf,
		dbLayer: &dbLayer{db: db},
		srv:     srv,
		l:       l,
	}

	l.WithField("port", conf.Port).Infoln("starting the server...")
	go func() {
		a.handleRequests(conf, srv, router)
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-term

	l.Infoln("shutting down the server...")
	if err := srv.Shutdown(context.Background()); err != nil {
		l.WithError(err).Errorln("server shutdown failed")
	}

	l.Infoln("closing all db connections...")
	db.Close()

	l.Infoln("goodbye")
}

func (a *App) handleRequests(conf *Config, srv *http.Server, router *mux.Router) {
	router.Handle("/ping", http.HandlerFunc(a.pingEndpointHandler)).Methods("GET")
	router.Handle("/tags", http.HandlerFunc(a.GetTagsHandler)).Methods("GET")
	router.Handle("/launch-commands", http.HandlerFunc(a.GetLaunchCommandsHandler)).Methods("GET")
	srv.ListenAndServe()
}
