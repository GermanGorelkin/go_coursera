package main

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"strconv"
	"fmt"
	"encoding/json"
	"sort"
	"strings"
	"io/ioutil"
	"encoding/xml"
	"time"
)

const (
	AccessToken = `Good token`
)

type UserModel struct {
	Id     int `xml:"id"`
	FirstName   string `xml:"first_name"`
	LastName   string `xml:"last_name"`
	Age    int `xml:"age"`
	About  string `xml:"about"`
	Gender string `xml:"gender"`
}

type Users struct {
	List    []UserModel `xml:"row"`
}

func (v *UserModel) Name() string{
	return fmt.Sprintf("%s %s", v.FirstName, v.LastName)
}

func OrderByID(users []UserModel) []UserModel{
	sort.Slice(users, func(i, j int) bool {
		return users[i].Id < users[j].Id
	})
	return users
}

func OrderByAge(users []UserModel) []UserModel{
	sort.Slice(users, func(i, j int) bool {
		return users[i].Age < users[j].Age
	})
	return users
}

func OrderByName(users []UserModel) []UserModel{
	sort.Slice(users, func(i, j int) bool {
		return strings.Compare(users[i].Name(), users[j].Name()) < 0
	})
	return users
}

func Limit(users []UserModel, limit int, offset int) ([]UserModel, error) {
	if offset >= len(users) || offset+limit >= len(users) {
		return users, fmt.Errorf("limit or offset > users")
	}
	return users[offset : offset+limit], nil
}

func (v *Users) LoadXml(filename string){
	xmlData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	err = xml.Unmarshal(xmlData, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
}

func (v *Users) Search(s string) []UserModel {
	users := make([]UserModel, 0)
	for _, u := range v.List {
		if strings.Contains(u.About, s) || strings.Contains(u.Name(), s) {
			users = append(users, u)
		}
	}
	return users
}

type TestCase struct {
	Request SearchRequest
	AccessToken string
	Result  *SearchResponse
	IsError bool
	Error error
}

func SearchServer (w http.ResponseWriter, r *http.Request) {
	at := r.Header.Get("AccessToken")
	if AccessToken != at {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var err error

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy, err := strconv.Atoi(r.URL.Query().Get("order_by"))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u := new(Users)
	u.LoadXml("dataset.xml")

	users := u.Search(query)

	switch orderField {
	case "Id":
		{
			users = OrderByID(users)
		}
	case "Age":
		{
			users = OrderByAge(users)
		}
	case "Name":
		{
			users = OrderByAge(users)
		}
	default:
		{
			errBody, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})

			w.WriteHeader(http.StatusBadRequest)
			w.Write(errBody)
			return
		}
	}


	switch orderBy {
	case -1:
		{
			break
		}
	case 0:
		{
			break
		}
	case 1:
		{
			break
		}
	case 2:
		{
			time.Sleep(2 * time.Second)
		}
	case 3:
		{
			return
		}
	case 4:
		{
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case 5:
		{
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case 6:
		{

			v, _ := json.Marshal(User{
				Id:     1,
				Name:   "name",
				About:  "name",
				Age:    1,
				Gender: "name",
			})

			w.WriteHeader(http.StatusBadRequest)
			w.Write(v)
			return
		}
	case 7:
		{
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	if orderBy == 1 {
		time.Sleep(time.Second)
	}

	users, err = Limit(users, limit, offset)

	result := make([]User, len(users))
	for i, u := range users {
		result[i] = User{
			Id:     u.Id,
			Name:   u.Name(),
			About:  u.About,
			Age:    u.Age,
			Gender: u.Gender,
		}
	}
	resBody, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(resBody)
}

func TestSearchClientFindUsersLimit(t *testing.T) {
	cases := []TestCase{
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      -1,
				Offset:     5,
				OrderBy:    0,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("limit must be > 0"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      5,
				Offset:     -1,
				OrderBy:    0,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("offset must be > 0"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     0,
				OrderBy:    0,
				OrderField: "Id",
				Query:      "",
			},
			IsError: false,
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     10,
				OrderBy:    0,
				OrderField: "Id",
				Query:      "",
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		c := &SearchClient{
			URL: ts.URL,
			AccessToken: item.AccessToken,
		}
		_, err := c.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
	}

	ts.Close()
}

func TestSearchClientFindUsersBadStatus(t *testing.T) {
	cases := []TestCase{
		TestCase{
			AccessToken: "opss",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				OrderBy:    0,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("Bad AccessToken"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				OrderBy:    0,
				OrderField: "Fail",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("StatusBadRequest"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     10,
				OrderBy:    2,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("Timeout"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     10,
				OrderBy:    4,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("StatusInternalServerError"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     10,
				OrderBy:    5,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("StatusBadRequest"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     10,
				OrderBy:    6,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("StatusBadRequest"),
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      30,
				Offset:     10,
				OrderBy:    7,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
			Error:   fmt.Errorf("StatusBadRequest"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		c := &SearchClient{
			URL: ts.URL,
			AccessToken: item.AccessToken,
		}
		_, err := c.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
	}

	ts.Close()
}

func TestSearchClientFindUsersUrl(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	c := &SearchClient{
		URL:        "",
		AccessToken: AccessToken,
	}
	c.FindUsers(SearchRequest{
		Limit:      30,
		Offset:     10,
		OrderBy:    0,
		OrderField: "Id",
		Query:      "",
	})


	ts.Close()
}