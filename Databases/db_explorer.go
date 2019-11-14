package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, msg)
}

func sendAnswer(w http.ResponseWriter, content map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(map[string]interface{}{
		"response": content,
	})
	if err != nil {
		panic(err)
	}

	w.Write(body)
}

// Handler ...
type Handler struct {
	Name  string
	Agent DbAgent
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/") { // sometimes url starts from "/"
		path = path[1:]
	}

	url := strings.Split(path, "/")
	table := url[0]
	if table == "" {
		tables := h.Agent.GetTables()
		sort.Strings(tables)
		sendAnswer(w, map[string]interface{}{
			"tables": tables,
		})
		return
	} else if _, ok := h.Agent.Schema[table]; !ok {
		sendError(w, "unknown table", http.StatusNotFound)
		return
	}

	switch len(url) {
	case 1:
		r.ParseForm()
		if r.Method == http.MethodGet {
			docs, e := h.Agent.GetRows(
				table, r.Form.Get("limit"),
				r.Form.Get("offset"),
			)
			if e != nil {
				sendError(w, e.Error(), http.StatusInternalServerError)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"records": docs,
			})
		} else if r.Method == http.MethodPut {
			insertID, e := h.Agent.NewRow(table, r.Form)
			if e != nil {
				sendError(w, e.Error(), http.StatusInternalServerError)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"user_id": insertID,
			})
		}
	case 2:
		id := url[1]

		switch r.Method {
		case http.MethodGet:
			doc, e := h.Agent.GetRow(table, id)
			if e != nil {
				sendError(w, "record not found", http.StatusNotFound)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"record": doc,
			})
		case http.MethodPost:
			r.ParseForm()
			updateID, e := h.Agent.UpdateRow(table, id, r.Form)
			if e != nil {
				sendError(w, e.Error(), http.StatusInternalServerError)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"updated": updateID,
			})
		case http.MethodDelete:
			deleteID, e := h.Agent.DeleteRow(table, id)
			if e != nil {
				sendError(w, e.Error(), http.StatusInternalServerError)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"deleted": deleteID,
			})
		}
	default:
		sendError(w, "too many endpoints in url", http.StatusBadRequest)
	}
}

// NewDbExplorer make handler for hw6 http service
func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	agent := DbAgent{db, nil}

	if e := agent.ReadDbSchema(); e != nil {
		return nil, e
	}

	return &Handler{"CRUD-service", agent}, nil
}
