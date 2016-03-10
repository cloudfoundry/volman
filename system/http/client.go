package http

import os_http "net/http"

//go:generate counterfeiter -o ../../volmanfakes/fake_http_client.go . Client

type Client interface {
	Do(req *os_http.Request) (resp *os_http.Response, err error)
}

func NewClientFrom(httpClient *os_http.Client) Client {
	return &systemHttpClient{httpClient}
}

func NewClient() Client {
	return &systemHttpClient{&os_http.Client{}}
}

type systemHttpClient struct {
	Delegate *os_http.Client
}

func (client *systemHttpClient) Do(req *os_http.Request) (resp *os_http.Response, err error) {
	return client.Delegate.Do(req)
}
