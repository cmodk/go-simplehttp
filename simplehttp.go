package simplehttp

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	net_url "net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type SimpleHttp struct {
	server         string
	debug          bool
	static_headers map[string]string
	logger         *logrus.Logger
}

func New(s string, lg *logrus.Logger) SimpleHttp {
	sh := SimpleHttp{
		server:         s,
		logger:         lg,
		debug:          false,
		static_headers: make(map[string]string),
	}

	return sh
}

func (sh *SimpleHttp) SetDebug(d bool) {
	sh.debug = d
}

func (sh *SimpleHttp) AddHeader(k string, v string) {
	sh.static_headers[k] = v
}

func (sh *SimpleHttp) SetBearerAuth(key string) {
	sh.static_headers["Authorization"] = "Bearer " + key
}

func (sh *SimpleHttp) SetBasicAuth(username string, password string) {
	auth := username + ":" + password
	basic := base64.StdEncoding.EncodeToString([]byte(auth))

	sh.static_headers["Authorization"] = "Basic " + basic
}

func (sh *SimpleHttp) set_headers(r *http.Request) {
	for k, v := range sh.static_headers {
		r.Header.Add(k, v)
	}
}

func (sh *SimpleHttp) Get(url string) (string, error) {
	url = sh.server + url

	client := &http.Client{}

	if sh.debug {
		log.Printf("GET: %s\n", url)
	}

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	sh.set_headers(r)
	resp, err := client.Do(r)
	if err != nil {
		return "", nil
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.New(fmt.Sprintf("GET error: %d:\nBody: %s\n", resp.StatusCode, string(body)))
		sh.logger.WithField("error", err)
		return string(body), err
	}

	return string(body), nil

}

func (sh *SimpleHttp) Post(url string, data interface{}) (string, error) {
	url = sh.server + url

	client := &http.Client{}

	if sh.debug {
		log.Printf("POST: %s\n", url)
	}

	extra_headers := make(map[string]string)
	var reader io.Reader
	switch t := data.(type) {
	case net_url.Values:
		reader = strings.NewReader(t.Encode())
		extra_headers["Content-Type"] = "application/x-www-form-urlencoded"
	default:
		return "", errors.New(fmt.Sprintf("Unhandled type for POST: %t\n", t))
	}

	r, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return "", err
	}

	for k, h := range extra_headers {
		r.Header.Add(k, h)
	}
	sh.set_headers(r)

	req, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("Could not dump request: %s\n", err.Error)
	} else {
		log.Printf("Request: %s\n", req)
	}
	resp, err := client.Do(r)
	if err != nil {
		return "", nil
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.New(fmt.Sprintf("GET error: %d:\nBody: %s\n", resp.StatusCode, string(body)))
		sh.logger.WithField("error", err)
		return string(body), err
	}

	return string(body), nil

}
