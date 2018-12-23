package fetch

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"
)

func Head(url string) (resp *http.Response, err error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
	}

	Client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transCfg,
	}
	resp, err = Client.Head(url)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func Get(url string) (content []byte, err error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
	}

	Client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transCfg,
	}
	resp, err := Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)

	return bytes, nil
}
