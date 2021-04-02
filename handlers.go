package main

import (
	"bytes"
	"encoding/json"
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
	for i := 0; i < 30; i++ {
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
		lc(ctx).Fatal("failed to open database")
		os.Exit(1)
	}
	os.Remove(bakName) // TODO handle error
	lc(ctx).Debugln("db update finished")
}

type tagsResp struct {
	Tags []string `json:"tags"`
}

// GetTagsHandler returns JSON list of tags
func (a *App) GetTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a.dbLayer.Lock()
	defer a.dbLayer.Unlock()

	rows, err := a.dbLayer.db.Query("SELECT name from tag_alias")
	if err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	resp := tagsResp{Tags: make([]string, 0, 1000)}
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			lc(ctx).Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		resp.Tags = append(resp.Tags, s)
	}

	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(resp); err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadBuf.Bytes())
}

type lauchCommandsResp struct {
	LaunchCommands []string `json:"launch_commands"`
}

// GetLaunchCommandsHandler returns JSON list of tags
func (a *App) GetLaunchCommandsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a.dbLayer.Lock()
	defer a.dbLayer.Unlock()

	rows, err := a.dbLayer.db.Query("SELECT launchCommand from game")
	if err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	resp := lauchCommandsResp{LaunchCommands: make([]string, 0, 100000)}
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			lc(ctx).Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		resp.LaunchCommands = append(resp.LaunchCommands, s)
	}

	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(resp); err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadBuf.Bytes())
}

// GetMasterDB returns the master DB
func (a *App) GetMasterDB(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a.dbLayer.Lock()
	dbFile, err := ioutil.ReadFile(dbName)
	a.dbLayer.Unlock()

	if err != nil {
		lc(ctx).Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", "attachment; filename=flashpoint-master-db.sqlite3")
	w.Write(dbFile)
}
