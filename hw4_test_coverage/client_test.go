package main

import (
	"net/http"
	"strconv"
	"fmt"
	"encoding/json"
	"sort"
	"strings"
	"io/ioutil"
	"encoding/xml"
	"net/http/httptest"
	"testing"
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

func OrderByID(users []UserModel, orderBy int) []UserModel{
	sort.Slice(users, func(i, j int) bool {
		if orderBy == 1 {
			return users[i].Id < users[j].Id
		} else {
			return users[i].Id > users[j].Id
		}
	})
	return users
}

func OrderByAge(users []UserModel, orderBy int) []UserModel{
	sort.Slice(users, func(i, j int) bool {
		if orderBy == 1 {
			return users[i].Age < users[j].Age
		}else {
			return users[i].Age > users[j].Age
		}
	})
	return users
}

func OrderByName(users []UserModel, orderBy int) []UserModel{
	sort.Slice(users, func(i, j int) bool {
		if orderBy == 1 {
			return strings.Compare(users[i].Name(), users[j].Name()) < 0
		} else{
			return strings.Compare(users[i].Name(), users[j].Name()) > 0
		}
	})
	return users
}

func Limit(users []UserModel, limit int, offset int) ([]UserModel, error) {
	if offset >= len(users) {
		return users, fmt.Errorf("offset > users")
	}
	if offset+limit > len(users) {
		return users[offset : len(users)-1], nil
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

	//!!!
	switch orderBy {
	case 2: // fmt.Errorf("timeout for %s", searcherParams.Encode())
		{
			time.Sleep(2 * time.Second)
		}
	case 3:
		{  // fmt.Errorf("SearchServer fatal error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case 4:
		{ // fmt.Errorf("unknown bad request error: %s", errResp.Error)
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
	case 5:
		{ // fmt.Errorf("cant unpack result json: %s", err)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	//

	// -1 по убыванию, 0 как встретилось, 1 по возрастанию
	if orderBy != 0 {
		switch orderField {
		case "Id":
			{
				users = OrderByID(users, orderBy)
			}
		case "Age":
			{
				users = OrderByAge(users, orderBy)
			}
		case "Name":
			{
				users = OrderByName(users, orderBy)
			}
		default:
			{
				errBody, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})

				w.WriteHeader(http.StatusBadRequest)
				w.Write(errBody)
				return
			}
		}

	}

	users, err = Limit(users, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

	w.Write(resBody)
}

// testing

type TestCase struct {
	Request SearchRequest
	AccessToken string
	Result  *SearchResponse
	IsError bool
	Error error
}

func TestSearchClient_FindUsers_Limit(t *testing.T) {
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

func TestSearchClient_FindUsers_Offset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	c := &SearchClient{
		URL:         ts.URL,
		AccessToken: AccessToken,
	}
	_, err := c.FindUsers(SearchRequest{
		Limit:      5,
		Offset:     1000,
		OrderBy:    0,
		OrderField: "Id",
		Query:      "",
	})

	if err == nil {
		t.Errorf("")
	}

	ts.Close()
}

func TestSearchClient_FindUsers_AccessToken(t *testing.T) {
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

func TestSearchClient_FindUsers_URL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	c := &SearchClient{
		URL:         "",
		AccessToken: AccessToken,
	}
	_, err := c.FindUsers(SearchRequest{
		Limit:      30,
		Offset:     10,
		OrderBy:    0,
		OrderField: "Id",
		Query:      "",
	})

	if err == nil {
		t.Errorf("")
	}

	ts.Close()
}

func TestSearchClient_FindUsers_BadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	c := &SearchClient{
		URL:         ts.URL,
		AccessToken: AccessToken,
	}
	_, err := c.FindUsers(SearchRequest{
		Limit:      30,
		Offset:     10,
		OrderBy:    1,
		OrderField: "UID",
		Query:      "",
	})

	if err == nil {
		t.Errorf("")
	}

	ts.Close()
}

func TestSearchClient_FindUsers_Other(t *testing.T) {
	cases := []TestCase{
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      1,
				Offset:     1,
				OrderBy:    2,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      1,
				Offset:     1,
				OrderBy:    3,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      1,
				Offset:     1,
				OrderBy:    4,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
		},
		TestCase{
			AccessToken: AccessToken,
			Request: SearchRequest{
				Limit:      1,
				Offset:     1,
				OrderBy:    5,
				OrderField: "Id",
				Query:      "",
			},
			IsError: true,
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