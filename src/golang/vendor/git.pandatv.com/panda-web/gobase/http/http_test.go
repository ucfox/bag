package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
	rsp, err := GetAsString("http://facebook.com")
	fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	if rsp == "" {
		t.Fatal("rsp nil")
	}
	fmt.Println(rsp)
}

func TestGetString(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	rsp, err := GetAsString(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if rsp == "" {
		t.Fatal("rsp nil")
	}
	fmt.Println(rsp)
}

func TestPostFormString(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	params := url.Values{}
	params.Add("i", "1")
	params.Add("s", "a")
	rsp, err := PostFormAsString(s.URL, params)
	if err != nil {
		t.Fatal(err)
	}
	if rsp == "" {
		t.Fatal("rsp nil")
	}
	fmt.Println(rsp)
}

func TestPostString(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	params := url.Values{}
	params.Add("i", "1")
	params.Add("s", "a")
	rsp, err := PostAsString(s.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	if rsp == "" {
		t.Fatal("rsp nil")
	}
	fmt.Println(rsp)
}

func TestDoString(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	r, err := http.NewRequest("get", s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	rsp, err := DoAsString(r)
	if err != nil {
		t.Fatal(err)
	}
	if rsp == "" {
		t.Fatal("rsp nil")
	}
	fmt.Println(rsp)
}

func TestGetJson(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	var d Result
	err := GetAsJson(s.URL+"?i=1&s=a", &d)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", d)
	if d.Errno == 0 || d.D.I != 1 || d.D.S != "a" {
		t.Fatal("rsp json err")
	}
}

func TestPostFormJson(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	params := url.Values{}
	params.Add("i", "1")
	params.Add("s", "a")
	var d Result
	err := PostFormAsJson(s.URL, params, &d)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", d)
	if d.Errno == 0 || d.D.I != 1 || d.D.S != "a" {
		t.Fatal("rsp json err")
	}
}

func TestPostJson(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	params := url.Values{}
	params.Add("i", "1")
	params.Add("s", "a")
	var d Result
	err := PostAsJson(s.URL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()), &d)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", d)
	if d.Errno == 0 || d.D.I != 1 || d.D.S != "a" {
		t.Fatal("rsp json err")
	}
}

func TestDoJson(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(httpTest))
	r, err := http.NewRequest("get", s.URL, strings.NewReader("i=1&s=a"))
	if err != nil {
		t.Fatal(err)
	}
	var d Result
	err = DoAsJson(r, &d)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", d)
	if d.Errno == 0 {
		t.Fatal("rsp json err")
	}
}

type Result struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	D      Data   `json:"data"`
}

type Data struct {
	I int    `json:"i"`
	S string `json:"s"`
}

func httpTest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	_i := r.Form.Get("i")
	s := r.Form.Get("s")
	var i int
	if _i != "" {
		i, _ = strconv.Atoi(_i)
	}
	d := &Result{
		Errno:  1,
		Errmsg: "success",
		D: Data{
			I: i,
			S: s,
		},
	}
	data, _ := json.Marshal(d)
	w.Write(data)
}
