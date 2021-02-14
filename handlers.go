package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func (a *App) pingEndpointHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

// UpdateDBHandler handles DB update from the god himself
func (a *App) UpdateDBHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lc(ctx).Debugln("db update started")

	r.ParseMultipartForm(10 << 28)
	file, handler, err := r.FormFile("db")
	if err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	lc(ctx).Debugf("received file size: %dB", handler.Size)

	lc(ctx).Debugln("reading the payload")
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	lc(ctx).Debugln("saving payload to a tmp db file")
	tmpDBName := "tmp.db"
	if err := ioutil.WriteFile(tmpDBName, buf, 0644); err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	lc(ctx).Debugln("locking db access")
	a.dbLayer.Lock()
	defer a.dbLayer.Unlock()
	a.dbLayer.db.Close()
	time.Sleep(100 * time.Millisecond)

	lc(ctx).Debugln("updating db")
	bakName := dbName + ".bak"
	os.Rename(dbName, bakName) // TODO handle error

	if err := os.Rename(tmpDBName, dbName); err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		os.Rename(bakName, dbName) // TODO handle error
		return
	}

	var openErr error
	for i := 0; i < 10; i++ {
		lc(ctx).Debugln("opening new db")
		a.dbLayer.db, openErr = openDatabase(a.l)
		if openErr == nil {
			break
		}
		lc(ctx).Error("unable to open database, retrying after 100ms...")
		time.Sleep(100 * time.Millisecond)
	}
	if openErr != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		os.Rename(bakName, dbName) // TODO handle error
		return
	}
	os.Remove(bakName) // TODO handle error
	lc(ctx).Debugln("db update finished")
}
