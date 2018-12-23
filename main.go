package main

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"goWhatweb/until"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	sortPairs, webdata := until.ParseCmsDataFromFile("cms.json")

	// 开始并发相关
	t1 := time.Now()
	domain := "https://x.hacking8.com" // 域名
	count := make(chan int, 200)       // 并发总量
	quite := make(chan int)            // 控制退出

	for _, v := range sortPairs {

		count <- 1
		go worker(domain, v.Path, webdata, &count, &quite)

	}
	for {

		if len(count) == 0 {
			quite <- 1
		}
		select {
		case <-quite:
			elapsed := time.Since(t1)
			fmt.Println("elapsed:", elapsed)
			return
		}
		//time.Sleep(time.Second)
	}
}

func worker(domain, path string, webdata map[string][]until.Singcms, count *chan int, quite *chan int) {
	domain_url := domain + path

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
	}
	Client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transCfg,
	}
	resp, err := Client.Head(domain_url)
	if err != nil {
		log.Println(err)
		<-*count
		return
	}
	if resp.StatusCode != 200 {
		<-*count
		return
	}

	response, err := Client.Get(domain_url)

	if err != nil {
		log.Println(err)
		<-*count
		return
	}

	defer response.Body.Close()

	if response.StatusCode == 404 {
		<-*count
		return
	}
	bytes, _ := ioutil.ReadAll(response.Body)

	cmsinfos := webdata[path]
	for _, cmsinfo := range cmsinfos {
		option := cmsinfo.Option
		if option == "keyword" {
			if strings.Contains(string(bytes), cmsinfo.Content) {
				fmt.Println(cmsinfo)
				*quite <- 1
				break
			}
		} else if option == "md5" {
			md5str := fmt.Sprintf("%x", md5.Sum(bytes))
			if md5str == cmsinfo.Content {
				fmt.Println(cmsinfo)
				*quite <- 1
				break
			}

		}
	}
	<-*count
	return
}
