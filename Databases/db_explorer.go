package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"crud"
)

func sendError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, err)
}

func sendAnswer(w http.ResponseWriter, content map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(map[string]interface{}{
		"response": content,
	})
	// should change handling
	if err != nil {
		panic(err)
	}

	w.Write(body)
}

func parseBody(reqBody io.ReadCloser) (map[string]interface{}, error) {
	defer reqBody.Close()
	body, err := ioutil.ReadAll(reqBody)
	if err != nil {
		return nil, err
	}

	doc := make(map[string]interface{})
	err = json.Unmarshal(body, &doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// Handler ...
type Handler struct {
	Name  string
	Agent crud.Agent
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// sometimes url starts from "/"
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	url := strings.Split(path, "/")
	table := url[0]
	if table == "" {
		tables := h.Agent.GetTables()

		sendAnswer(w, map[string]interface{}{
			"tables": tables,
		})
		return
	} else if _, ok := h.Agent.Schema[table]; !ok {
		sendError(w, errors.New("unknown table"), http.StatusNotFound)
		return
	}

	switch len(url) {
	case 1:
		if r.Method == http.MethodGet {
			r.ParseForm()
			docs, e := h.Agent.GetRows(
				table, r.Form.Get("limit"),
				r.Form.Get("offset"),
			)
			if e != nil {
				sendError(w, e, http.StatusInternalServerError)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"records": docs,
			})
		} else if r.Method == http.MethodPut {
			doc, err := parseBody(r.Body)
			if err != nil {
				sendError(w, err, http.StatusBadRequest)
				break
			}

			prKey, insertID, err := h.Agent.NewRow(table, doc)
			if err != nil {
				sendError(w, err, http.StatusInternalServerError)
				break
			}

			res := make(map[string]interface{}, 1)
			res[prKey] = insertID
			sendAnswer(w, res)
		}
	case 2:
		id := url[1]

		switch r.Method {
		case http.MethodGet:
			doc, e := h.Agent.GetRow(table, id)
			if e != nil {
				sendError(w, errors.New("record not found"), http.StatusNotFound)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"record": doc,
			})
		case http.MethodPost:
			doc, err := parseBody(r.Body)
			if err != nil {
				sendError(w, err, http.StatusBadRequest)
				break
			}

			updateID, e := h.Agent.UpdateRow(table, id, doc)
			if e != nil {
				sendError(w, e, http.StatusBadRequest)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"updated": updateID,
			})
		case http.MethodDelete:
			deleteID, e := h.Agent.DeleteRow(table, id)
			if e != nil {
				sendError(w, e, http.StatusInternalServerError)
				break
			}

			sendAnswer(w, map[string]interface{}{
				"deleted": deleteID,
			})
		}
	default:
		sendError(w, errors.New("too many endpoints in url"), http.StatusBadRequest)
	}
}

// NewDbExplorer make handler for hw6 http service
func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	agent := crud.Agent{db, nil}

	if e := agent.ReadDbSchema(); e != nil {
		return nil, e
	}

	return &Handler{"CRUD-service", agent}, nil
}
