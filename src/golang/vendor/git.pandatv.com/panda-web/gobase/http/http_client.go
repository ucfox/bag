package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	*http.Client
}

var defaultClient = &Client{
	&http.Client{
		Timeout: time.Second,
	},
}

func NewClient(c *http.Client) *Client {
	return &Client{c}
}

func Get(url string) (*http.Response, error) {
	return defaultClient.Get(url)
}

func (c *Client) Get(url string) (*http.Response, error) {
	return c.Client.Get(url)
}

func PostForm(url string, params url.Values) (*http.Response, error) {
	return defaultClient.PostForm(url, params)
}

func (c *Client) PostForm(url string, params url.Values) (*http.Response, error) {
	return c.Client.PostForm(url, params)
}

func Post(url string, boydType string, body io.Reader) (*http.Response, error) {
	return defaultClient.Post(url, boydType, body)
}

func (c *Client) Post(url string, boydType string, body io.Reader) (*http.Response, error) {
	return c.Client.Post(url, boydType, body)
}

func Do(req *http.Request) (*http.Response, error) {
	return defaultClient.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

func GetAsString(url string) (string, error) {
	return toString(Get(url))
}

func (c *Client) GetAsString(url string) (string, error) {
	return toString(c.Get(url))
}

func GetAsJson(url string, v interface{}) error {
	rsp, err := Get(url)
	return toJson(rsp, err, v)
}

func (c *Client) GetAsJson(url string, v interface{}) error {
	rsp, err := c.Get(url)
	return toJson(rsp, err, v)
}

func PostFormAsString(url string, params url.Values) (string, error) {
	return toString(PostForm(url, params))
}

func (c *Client) PostFormAsString(url string, params url.Values) (string, error) {
	return toString(c.PostForm(url, params))
}

func PostFormAsJson(url string, params url.Values, v interface{}) error {
	rsp, err := PostForm(url, params)
	return toJson(rsp, err, v)
}

func (c *Client) PostFormAsJson(url string, params url.Values, v interface{}) error {
	rsp, err := c.PostForm(url, params)
	return toJson(rsp, err, v)
}

func PostAsString(url, bodyType string, body io.Reader) (string, error) {
	return toString(Post(url, bodyType, body))
}

func (c *Client) PostAsString(url, bodyType string, body io.Reader) (string, error) {
	return toString(c.Post(url, bodyType, body))
}

func PostAsJson(url, bodyType string, body io.Reader, v interface{}) error {
	rsp, err := Post(url, bodyType, body)
	return toJson(rsp, err, v)
}

func (c *Client) PostAsJson(url, bodyType string, body io.Reader, v interface{}) error {
	rsp, err := c.Post(url, bodyType, body)
	return toJson(rsp, err, v)
}

func DoAsString(r *http.Request) (string, error) {
	return toString(Do(r))
}

func (c *Client) DoAsString(r *http.Request) (string, error) {
	return toString(c.Do(r))
}

func DoAsJson(r *http.Request, v interface{}) error {
	rsp, err := Do(r)
	return toJson(rsp, err, v)
}

func (c *Client) DoAsJson(r *http.Request, v interface{}) error {
	rsp, err := c.Do(r)
	return toJson(rsp, err, v)
}

func toString(rsp *http.Response, err error) (string, error) {
	if rsp != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	if rsp.StatusCode != 200 {
		return "", errors.New(rsp.Status)
	}
	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func toJson(rsp *http.Response, err error, v interface{}) error {
	if rsp != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
		return err
	}
	if rsp.StatusCode != 200 {
		return errors.New(rsp.Status)
	}
	err = json.NewDecoder(rsp.Body).Decode(v)
	if err != nil {
		return err
	}
	return nil
}
