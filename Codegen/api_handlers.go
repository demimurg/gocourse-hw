
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type srvRes struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
}
 
func SendError(w http.ResponseWriter, status int, mes string) {
	w.WriteHeader(status)
	paramErr, _ := json.Marshal(srvRes{Error: mes})
	w.Write(paramErr)
}


func (params *ProfileParams) ParseAndValidate(reqValues url.Values) error {
	params.Login = reqValues.Get("login")
	
	if params.Login == "" {
		return fmt.Errorf("login must me not empty")
	}
	

	return nil
}

func (params *CreateParams) ParseAndValidate(reqValues url.Values) error {
	params.Login = reqValues.Get("login")
	
	if params.Login == "" {
		return fmt.Errorf("login must me not empty")
	}
	
	if len(params.Login) < 10 {
		return fmt.Errorf("login len must be >= 10")
	}
	

	params.Name = reqValues.Get("full_name")
	

	params.Status = reqValues.Get("status")
	if params.Status == "" {
		params.Status = "user"
	}
	  
	if params.Status != "user" && params.Status != "moderator" && params.Status != "admin"  {
		return fmt.Errorf("status must be one of [user, moderator, admin]")
	}
	

	var err error
	params.Age, err = strconv.Atoi(reqValues.Get("age"))
	if err != nil {
		return fmt.Errorf("age must be int")
	}
	
	if params.Age < 0 {
		return fmt.Errorf("age must be >= 0")
	}
	if params.Age > 128 {
		return fmt.Errorf("age must be <= 128")
	}
	

	return nil
}

func (params *OtherCreateParams) ParseAndValidate(reqValues url.Values) error {
	params.Username = reqValues.Get("username")
	
	if params.Username == "" {
		return fmt.Errorf("username must me not empty")
	}
	
	if len(params.Username) < 3 {
		return fmt.Errorf("username len must be >= 3")
	}
	

	params.Name = reqValues.Get("account_name")
	

	params.Class = reqValues.Get("class")
	if params.Class == "" {
		params.Class = "warrior"
	}
	  
	if params.Class != "warrior" && params.Class != "sorcerer" && params.Class != "rouge"  {
		return fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")
	}
	

	var err error
	params.Level, err = strconv.Atoi(reqValues.Get("level"))
	if err != nil {
		return fmt.Errorf("level must be int")
	}
	
	if params.Level < 1 {
		return fmt.Errorf("level must be >= 1")
	}
	if params.Level > 50 {
		return fmt.Errorf("level must be <= 50")
	}
	

	return nil
}

func (srv *MyApi) HandlerProfile(w http.ResponseWriter, req *http.Request) {
	
	req.ParseForm()
	params := ProfileParams{}

	err := params.ParseAndValidate(req.Form)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := srv.Profile(req.Context(), params)

	if err == nil {
		w.WriteHeader(http.StatusOK)

		body, _ := json.Marshal(
			srvRes{Response: user},
		)
		w.Write(body)
	} else if e, ok := err.(ApiError); ok {
		SendError(w, e.HTTPStatus, e.Error())
	} else {
		SendError(w, http.StatusInternalServerError, err.Error())
	}
}

func (srv *MyApi) HandlerCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		SendError(w, http.StatusNotAcceptable, "bad method")
		return
	}
	
	if req.Header.Get("X-AUTH") != "100500" {
		SendError(w, http.StatusForbidden, "unauthorized")
		return
	}
	
	req.ParseForm()
	params := CreateParams{}

	err := params.ParseAndValidate(req.Form)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := srv.Create(req.Context(), params)

	if err == nil {
		w.WriteHeader(http.StatusOK)

		body, _ := json.Marshal(
			srvRes{Response: user},
		)
		w.Write(body)
	} else if e, ok := err.(ApiError); ok {
		SendError(w, e.HTTPStatus, e.Error())
	} else {
		SendError(w, http.StatusInternalServerError, err.Error())
	}
}

func (srv *OtherApi) HandlerCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		SendError(w, http.StatusNotAcceptable, "bad method")
		return
	}
	
	if req.Header.Get("X-AUTH") != "100500" {
		SendError(w, http.StatusForbidden, "unauthorized")
		return
	}
	
	req.ParseForm()
	params := OtherCreateParams{}

	err := params.ParseAndValidate(req.Form)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := srv.Create(req.Context(), params)

	if err == nil {
		w.WriteHeader(http.StatusOK)

		body, _ := json.Marshal(
			srvRes{Response: user},
		)
		w.Write(body)
	} else if e, ok := err.(ApiError); ok {
		SendError(w, e.HTTPStatus, e.Error())
	} else {
		SendError(w, http.StatusInternalServerError, err.Error())
	}
}

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
	case "/user/profile":
		srv.HandlerProfile(w, r)
	case "/user/create":
		srv.HandlerCreate(w, r)
	default:
		SendError(w, http.StatusNotFound, "unknown method")
	}
}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
	case "/user/create":
		srv.HandlerCreate(w, r)
	default:
		SendError(w, http.StatusNotFound, "unknown method")
	}
}
