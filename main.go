package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
)

type SubsonicResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version"`
	Type          string `json:"type"`
	ServerVersion string `json:"serverVersion"`
}

type SubsonicClient struct {
	client  *http.Client
	baseUrl string
	user    string
	salt    string
	token   string
}

func generateSalt() string {
	var corpus = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	// length is minimum 6, but let's use ten to start
	b := make([]rune, 10)
	for i := range b {
		b[i] = corpus[rand.Intn(len(corpus))]
	}
	return string(b)
}

func (s *SubsonicClient) Authenticate(password string) error {
	salt := generateSalt()
	h := md5.New()
	io.WriteString(h, password)
	io.WriteString(h, salt)
	s.salt = salt
	s.token = fmt.Sprintf("%x", h.Sum(nil))
	// Test authentication
	if !s.Ping() {
		return errors.New("Invalid authentication parameters!")
	}
	return nil
}

func (s *SubsonicClient) Request(method string, endpoint string, params map[string]string) ([]byte, error) {
	baseUrl, err := url.Parse(s.baseUrl)
	if err != nil {
		return nil, err
	}
	baseUrl.Path = path.Join(baseUrl.Path, "/rest/", endpoint)
	req, err := http.NewRequest(method, baseUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("f", "json")
	q.Add("v", "1.8.0")
	q.Add("c", "override-me")
	q.Add("u", s.user)
	q.Add("t", s.token)
	q.Add("s", s.salt)
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (s *SubsonicClient) Ping() bool {
	contents, err := s.Request("GET", "ping", nil)
	if err != nil {
		log.Fatal(err)
		return false
	}
	fmt.Printf("%s\n", contents)
	return true
}

func main() {
	client := SubsonicClient{
		client:  &http.Client{},
		baseUrl: "http://192.168.1.7:4040/",
		user:    "test",
	}
	err := client.Authenticate("blah")
	if err != nil {
		log.Fatal(err)
	}
}
