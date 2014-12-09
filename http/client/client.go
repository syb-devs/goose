package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

	ghttp "github.com/syb-devs/goose/http"
)

const (
	userAgent = "goose-api-go-client/0.1"
)

type dict map[string]string

type URLParams struct {
	Path  dict
	Query dict
}

type Service struct {
	BaseURL string
	client  *http.Client
	Objects *ObjectsService
	Buckets *BucketsService
}

func New(client *http.Client, baseURL string) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BaseURL: baseURL}
	s.Buckets = NewBucketsService(s)
	s.Objects = NewObjectsService(s)

	return s, nil
}

func (s *Service) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (s *Service) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := s.newRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

func (s *Service) do(req *http.Request) (*http.Response, error) {
	res, err := s.client.Do(req)
	if err != nil {
		return res, err
	}
	if httpErr := reponseError(res); httpErr != nil {
		return res, httpErr
	}
	return res, nil
}

func (s *Service) get(url string) (*http.Response, error) {
	return s.doRequest("GET", url, nil)
}

func (s *Service) delete(url string) (*http.Response, error) {
	return s.doRequest("DELETE", url, nil)
}

func (s *Service) url(path string, params *URLParams) (string, error) {
	if params == nil {
		return s.BaseURL + path, nil
	}
	for k, v := range params.Path {
		path = strings.Replace(path, "{"+k+"}", v, -1)
	}
	u, err := url.Parse(s.BaseURL + path)
	if err != nil {
		return "", err
	}
	up := url.Values{}
	for k, v := range params.Query {
		up.Add(k, v)
	}
	u.RawQuery = up.Encode()

	return u.String(), nil
}

func (s *Service) sendJSON(method, url string, data interface{}) (*http.Response, error) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)

	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return s.doRequest(method, url, b)
}

func (s *Service) getInto(url string, dest interface{}) error {
	res, err := s.get(url)
	if err != nil {
		return err
	}
	if err = decodeJSON(res.Body, dest); err != nil {
		return err
	}
	return nil
}

func reponseError(res *http.Response) *ghttp.Error {
	if res.StatusCode > 299 {
		httpErr := &ghttp.Error{}
		if err := decodeJSON(res.Body, httpErr); err == nil {
			return httpErr
		}
		return ghttp.NewError(res.StatusCode, "")
	}
	return nil
}

func decodeJSON(source io.Reader, dest interface{}) error {
	return json.NewDecoder(source).Decode(dest)
}

type ctxErr struct {
	ctx string
	err error
}

func (e *ctxErr) Error() string {
	return e.ctx + ": " + e.err.Error()
}

func newCtxErr(message string, err error) error {
	return &ctxErr{ctx: message, err: err}
}

func panicHandler() {
	if p := recover(); p != nil {
		print("recovering from panic: %v. \nstack trace: %s", p, debug.Stack())
	}
}
