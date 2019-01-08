package main

import (
	"bufio"
	"fmt"
	"goWhatweb/engine"
	"goWhatweb/until"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var ()

func main() {
	args := os.Args //获取用户输入的所有参数
	//
	if args == nil || len(args) != 2 {
		log.Fatalln("err:./goWhatweb test.txt")
		return
	}
	filename := args[1] //获取输入的第一个参数
	fmt.Println("Get filename:" + filename)

	fi, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer fi.Close()
	var domains []string

	br := bufio.NewReader(fi)
	for {
		s, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		str := string(s)
		domains = append(domains, str)

	}
	//domains := []string{"https://www.hacking8.com", "https://x.hacking8.com"}

	// 加载指纹
	sortPairs, webdata := until.ParseCmsDataFromFile("cms.json")
	var wg sync.WaitGroup

	// 开始并发相关
	t1 := time.Now()
	ResultChian := make(chan string)
	fmt.Println("Load url:", domains)
	for _, domain := range domains {
		go func(d string) {
			newWorker := engine.NewWorker(7, d, &wg, ResultChian)
			if !newWorker.Checkout() {
				return
			}
			newWorker.Start()
			for _, v := range sortPairs {
				tmp_job := engine.JobStruct{d, v.Path, webdata[v.Path]}
				//fmt.Println(tmp_job)
				newWorker.Add(tmp_job)
			}
		}(domain)
	}
	time.Sleep(time.Second * 2)
	go func() {
		for {
			r := <-ResultChian
			log.Println(r)
		}
	}()
	log.Println("初始化完成")

	wg.Wait()
	elapsed := time.Since(t1)
	fmt.Println("elapsed:", elapsed)

}
