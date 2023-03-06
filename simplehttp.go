package simplehttp

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	net_url "net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type SimpleHttp struct {
	server         string
	static_headers map[string]string
	logger         *logrus.Logger
	transport      *http.Transport
}

func New(s string, lg *logrus.Logger) SimpleHttp {
	sh := SimpleHttp{
		server:         s,
		logger:         lg,
		static_headers: make(map[string]string),
		transport:      &http.Transport{TLSClientConfig: &tls.Config{}},
	}

	return sh
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

func (sh *SimpleHttp) SetCustomCA(cert string) error {

	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	rootCAs.AddCert(c)

	config := &tls.Config{
		RootCAs: rootCAs,
	}

	sh.transport.TLSClientConfig = config

	return nil
}

func (sh *SimpleHttp) SkipTLSVerification() {
	sh.transport.TLSClientConfig.InsecureSkipVerify = true
}

func (sh *SimpleHttp) set_headers(r *http.Request) {
	for k, v := range sh.static_headers {
		r.Header.Add(k, v)
	}
}

func (sh *SimpleHttp) Get(url string) (string, error) {
	url = sh.server + url

	client := &http.Client{}

	client.Transport = sh.transport

	sh.logger.Printf("GET: %s\n", url)

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	sh.set_headers(r)
	resp, err := client.Do(r)
	if err != nil {
		return "", err
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
	return sh.PostPriv(url, data, "POST")
}

func (sh *SimpleHttp) Put(url string, data interface{}) (string, error) {
	return sh.PostPriv(url, data, "PUT")
}

func (sh *SimpleHttp) PostPriv(url string, data interface{}, method string) (string, error) {
	url = sh.server + url

	client := &http.Client{}

	sh.logger.Printf("%s: %s\n", method, url)

	extra_headers := make(map[string]string)
	var reader io.Reader
	switch t := data.(type) {
	case string:
		reader = strings.NewReader(t)
	case net_url.Values:
		reader = strings.NewReader(t.Encode())
		extra_headers["Content-Type"] = "application/x-www-form-urlencoded"
	default:
		//Default to json
		buf, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		reader = strings.NewReader(string(buf))
		extra_headers["Content-Type"] = "application/json"
	}

	r, err := http.NewRequest(method, url, reader)
	if err != nil {
		return "", err
	}

	for k, h := range extra_headers {
		r.Header.Add(k, h)
	}
	sh.set_headers(r)

	if sh.logger.Level == logrus.DebugLevel {
		req, err := httputil.DumpRequest(r, true)
		if err != nil {
			sh.logger.Errorf("Could not dump request: %s\n", err.Error)
		} else {
			sh.logger.Printf("Request: %s\n", req)
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.New(fmt.Sprintf("GET error: %d:\nBody: %s\n", resp.StatusCode, string(body)))
		sh.logger.WithField("error", err)
		return string(body), err
	}

	return string(body), nil

}
