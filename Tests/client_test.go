package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var users []User

func init() {
	type Сlient struct {
		FirstName string `xml:"first_name"`
		LastName  string `xml:"last_name"`
		ID        int    `xml:"id"`
		Age       int    `xml:"age"`
		About     string `xml:"about"`
		Gender    string `xml:"gender"`
	}

	type Сlients struct {
		List []Сlient `xml:"row"`
	}

	xmlFile, err := os.Open("dataset.xml")
	defer xmlFile.Close()
	if err != nil {
		log.Fatal(fmt.Sprintln("Open dataset.xml failed, error: ", err))
	}

	byteVal, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(fmt.Sprintln("Can't read xml file, error: ", err))
	}

	var dataset Сlients
	err = xml.Unmarshal(byteVal, &dataset)
	if err != nil {
		log.Fatal(fmt.Sprintln("Can't unmarshall dataset in Clients, error: ", err))
	}

	for _, cl := range dataset.List {
		users = append(users, User{
			Name:   cl.FirstName + " " + cl.LastName,
			Id:     cl.ID,
			Age:    cl.Age,
			About:  cl.About,
			Gender: cl.Gender,
		})
	}
}

func SearchServer(w http.ResponseWriter, req *http.Request) {
	const allowedToken = "666"

	param := req.URL.Query()
	var (
		accessToken = req.Header.Get("AccessToken")
		limit, _    = strconv.Atoi(param["limit"][0])
		offset, _   = strconv.Atoi(param["offset"][0])
		query       = param["query"][0]
		orderField  = param["order_field"][0]
		orderBy, _  = strconv.Atoi(param["order_by"][0])
	)

	if accessToken != allowedToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var requiredUsers []User
	for _, u := range users {
		haveQuery := strings.Contains(u.Name, query) ||
			strings.Contains(u.About, query)

		if haveQuery {
			requiredUsers = append(requiredUsers, u)
		}
	}

	if offset >= len(requiredUsers) {
		w.WriteHeader(http.StatusBadRequest)
		jsonErr, _ := json.Marshal(
			SearchErrorResponse{"Offset is bigger than number of docs"},
		)
		w.Write(jsonErr)
		return
	}

	correctOrder := false
	for _, order := range []int{-1, 0, 1} {
		if orderBy == order {
			correctOrder = true
			break
		}
	}
	if !correctOrder {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if orderBy != 0 {
		if !strings.Contains("Id, Age, Name", orderField) {
			w.WriteHeader(http.StatusBadRequest)

			jsonErr, _ := json.Marshal(
				SearchErrorResponse{"ErrorBadOrderField"},
			)
			w.Write(jsonErr)
			return
		}

		sort.SliceStable(requiredUsers, func(i, j int) bool {
			var relation bool
			switch orderField {
			case "Id":
				relation = requiredUsers[i].Id > requiredUsers[j].Id
			case "Age":
				relation = requiredUsers[i].Age > requiredUsers[j].Age
			case "Name", "":
				relation = requiredUsers[i].Name > requiredUsers[j].Name
			}

			if orderBy == -1 {
				relation = !relation
			}
			return relation
		})
	}

	requiredUsers = requiredUsers[offset:]

	if limit > 0 && limit < len(requiredUsers) {
		requiredUsers = requiredUsers[:limit]
	}

	usersJSON, _ := json.Marshal(requiredUsers)
	w.Write(usersJSON)
}

func TestBase(t *testing.T) {
	type TestCase struct {
		req     SearchRequest
		res     *SearchResponse
		isError bool
	}

	cases := []TestCase{
		{
			SearchRequest{
				Limit:      1,
				Query:      "Boyd",
				OrderField: "Name",
				OrderBy:    -1,
			},
			&SearchResponse{users[0:1], false},
			false,
		},
		{
			SearchRequest{
				Limit:      3,
				OrderField: "Id",
				OrderBy:    1,
			},
			&SearchResponse{
				[]User{users[34], users[33], users[32]},
				true,
			},
			false,
		},
		{
			SearchRequest{
				Limit:      10,
				Query:      "W",
				OrderField: "Id",
				OrderBy:    -1,
			},
			&SearchResponse{
				[]User{users[0], users[13], users[21], users[22]},
				false,
			},
			false,
		},
		{
			SearchRequest{OrderBy: -1, OrderField: "Djopa slona"},
			nil,
			true,
		},
		{
			SearchRequest{Limit: 40},
			&SearchResponse{users[0:25], true},
			false,
		},
		{
			SearchRequest{Offset: 40},
			nil,
			true,
		},
		{
			SearchRequest{OrderBy: -666},
			nil,
			true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	cli := SearchClient{AccessToken: "666", URL: ts.URL}
	defer ts.Close()

	for caseNum, item := range cases {
		cliRes, cliErr := cli.FindUsers(item.req)

		if cliErr != nil && !item.isError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, cliErr)
		}
		if cliErr == nil && item.isError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.res, cliRes) {
			t.Errorf("[%d] wrong result, \n\nexpected %#v, \n\ngot %#v", caseNum, item.res, cliRes)
		}
	}
}

func TestWithErrorCheck(t *testing.T) {
	type TestCase struct {
		req SearchRequest
		res *SearchResponse
		err error
	}

	cases := []TestCase{
		{
			SearchRequest{Limit: -1},
			nil,
			fmt.Errorf("limit must be > 0"),
		},
		{
			SearchRequest{Offset: -1},
			nil,
			fmt.Errorf("offset must be > 0"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	cli := SearchClient{AccessToken: "666", URL: ts.URL}
	defer ts.Close()

	for caseNum, item := range cases {
		cliRes, cliErr := cli.FindUsers(item.req)

		if !reflect.DeepEqual(cliErr, item.err) {
			t.Errorf("[%d] wrong error, \n\nexpected %#v, \n\ngot %#v", caseNum, item.err, cliErr)
		}
		if !reflect.DeepEqual(item.res, cliRes) {
			t.Errorf("[%d] wrong result, \n\nexpected %#v, \n\ngot %#v", caseNum, item.res, cliRes)
		}
	}
}

func TestAuth(t *testing.T) {
	type TestCase struct {
		token string
		err   error
	}

	cases := []TestCase{
		{"666", nil},
		{"333", fmt.Errorf("Bad AccessToken")},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	cli := SearchClient{URL: ts.URL}
	defer ts.Close()

	for caseNum, item := range cases {
		cli.AccessToken = item.token
		_, err := cli.FindUsers(SearchRequest{})
		if !reflect.DeepEqual(err, item.err) {
			t.Errorf("[%d] wrong error, \n\nexpected %#v, \n\ngot %#v", caseNum, item.err, err)
		}
	}
}

func TestBrokenServer(t *testing.T) {
	var (
		ts  *httptest.Server
		cli SearchClient
	)

	{
		ts = httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
		)
		cli = SearchClient{URL: ts.URL}

		_, err := cli.FindUsers(SearchRequest{})
		if !reflect.DeepEqual(err, fmt.Errorf("SearchServer fatal error")) {
			t.Errorf("Wrong handle of the Server Fatal Error: %#v", err)
		}
		ts.Close()
	}

	{
		ts = httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				io.WriteString(w, "{incorrect: json, djopa: slona}")
			}),
		)
		cli = SearchClient{URL: ts.URL}

		_, err := cli.FindUsers(SearchRequest{})
		if err == nil {
			t.Errorf("Incorrect json error not handled")
		}
		ts.Close()
	}

	{
		cli = SearchClient{}

		_, err := cli.FindUsers(SearchRequest{})
		if err == nil {
			t.Errorf("It can't be true")
		}
	}

	{
		ts = httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				time.Sleep(time.Second)
			}),
		)
		cli = SearchClient{URL: ts.URL}

		_, err := cli.FindUsers(SearchRequest{})
		if err == nil {
			t.Errorf("Must give timeout error: %#v", err)
		}
		ts.Close()
	}

}
