package req

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Request struct {
	method string
	url    string
	req    *http.Request
	body   io.Reader
	client *http.Client
}

func New(method, url string) (r *Request) {
	r = &Request{
		method: method,
		url:    url,
		client: &http.Client{},
	}
	return
}

func (r *Request) initReq() error {
	if r.req != nil {
		return nil
	}
	var err error
	r.req, err = http.NewRequest(r.method, r.url, r.body)
	return err
}

func (r *Request) Reset() {
	r.method = ""
	r.url = ""
	r.req = nil
	r.body = nil
}

func (r *Request) SetQuery(key, value string) error {
	if err := r.initReq(); err != nil {
		return err
	}
	q := r.req.URL.Query()
	q.Add(key, value)
	r.req.URL.RawQuery = q.Encode()
	return nil
}

func (r *Request) SetBody(body any) error {
	if b, ok := body.(io.Reader); ok {
		r.body = b
		return nil
	} else {
		jsonData, _ := json.Marshal(body)
		r.body = bytes.NewBuffer(jsonData)
	}

	if err := r.initReq(); err != nil {
		return err
	}
	r.req.Body = io.NopCloser(r.body)
	return nil
}

func (r *Request) SetHeader(key, value string) error {
	if err := r.initReq(); err != nil {
		return err
	}
	r.req.Header.Set(key, value)
	return nil
}

func (r *Request) Fetch() (*http.Response, error) {
	if err := r.initReq(); err != nil {
		return nil, err
	}
	return r.client.Do(r.req)
}
