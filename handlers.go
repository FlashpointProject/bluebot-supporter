package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (a *App) pingEndpointHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
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
