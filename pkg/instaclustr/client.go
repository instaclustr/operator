package instaclustr

import (
	v1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1"
	v2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2"
	"net/http"
	"time"
)

type ClientSet struct {
	username       string
	key            string
	serverHostname string
	httpClient     *http.Client
}

func NewClientSet(
	username string,
	key string,
	serverHostname string,
	timeout time.Duration,
) *ClientSet {
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{},
	}
	return &ClientSet{
		username:       username,
		key:            key,
		serverHostname: serverHostname,
		httpClient:     httpClient,
	}
}

func (c *ClientSet) V1() V1Interface {
	return v1.Client{
		Username:        c.username,
		Key:             c.key,
		ServerHostname:  c.serverHostname,
		HttpClient:      c.httpClient,
		OperatorVersion: OperatorVersion,
	}
}

func (c *ClientSet) V2() V2Interface {
	return v2.Client{
		Username:        c.username,
		Key:             c.key,
		ServerHostname:  c.serverHostname,
		HttpClient:      c.httpClient,
		OperatorVersion: OperatorVersion,
	}
}
