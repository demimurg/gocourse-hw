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

	if err != nil {
		sendError(
			w, errors.New("can't jsonify service answer"),
			http.StatusInternalServerError,
		)
		return
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

// Handler wrap crud.Agent, uses for routing
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

	var (
		url       = strings.Split(path, "/")
		table, id string
	)
	table = url[0]
	if len(url) > 1 {
		id = url[1]
	}

	if _, ok := h.Agent.Schema[table]; !ok && table != "" {
		sendError(w, errors.New("unknown table"), http.StatusNotFound)
		return
	}

	switch {

	// [GET "/"] get list of table names
	case r.Method == http.MethodGet && table == "":
		tables := h.Agent.GetTables()

		sendAnswer(w, tables)

	// [GET "/<table>"] get rows from the table
	case r.Method == http.MethodGet && len(url) == 1:
		r.ParseForm()

		records, e := h.Agent.GetRows(
			table, r.Form.Get("limit"),
			r.Form.Get("offset"),
		)
		if e != nil {
			sendError(w, e, http.StatusInternalServerError)
			break
		}

		sendAnswer(w, records)

	// [PUT "/<table>"] create a new row in the table
	case r.Method == http.MethodPut && len(url) == 1:
		body, err := parseBody(r.Body)
		if err != nil {
			sendError(w, err, http.StatusBadRequest)
			break
		}

		doc, err := h.Agent.Validate(table, "CREATE", body)
		if err != nil {
			sendError(w, err, http.StatusBadRequest)
			break
		}

		primaryKey, err := h.Agent.NewRow(table, doc)
		if err != nil {
			sendError(w, err, http.StatusInternalServerError)
			break
		}

		sendAnswer(w, primaryKey)

	// [GET "/<table>/<id>"] get the row by id
	case r.Method == http.MethodGet && len(url) == 2:
		record, e := h.Agent.GetRow(table, id)
		if e != nil {
			sendError(w, errors.New("record not found"), http.StatusNotFound)
			break
		}

		sendAnswer(w, record)

	// [POST "/<table>/<id>"] update the row by id
	case r.Method == http.MethodPost && len(url) == 2:
		body, err := parseBody(r.Body)
		if err != nil {
			sendError(w, err, http.StatusBadRequest)
			break
		}

		doc, err := h.Agent.Validate(table, "UPDATE", body)
		if err != nil {
			sendError(w, err, http.StatusBadRequest)
			break
		}

		updated, e := h.Agent.UpdateRow(table, id, doc)
		if e != nil {
			sendError(w, e, http.StatusBadRequest)
			break
		}

		sendAnswer(w, updated)

	// [DELETE "/<table>/<id>"]	delete the row from table
	case r.Method == http.MethodDelete && len(url) == 2:
		deleted, e := h.Agent.DeleteRow(table, id)
		if e != nil {
			sendError(w, e, http.StatusInternalServerError)
			break
		}

		sendAnswer(w, deleted)

	default:
		sendError(w, errors.New("uses wrong api"), http.StatusBadRequest)
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
