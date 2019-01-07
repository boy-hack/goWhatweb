package main

import (
	"fmt"
	"goWhatweb/engine"
	"goWhatweb/until"
	"log"
	"os"
	"time"
)

func main() {
	args := os.Args //获取用户输入的所有参数

	//如果用户没有输入,或参数个数不够,则调用该函数提示用户
	if args == nil || len(args) != 2 {
		log.Fatal("err:./goWhatweb https://x.hacking8.com")
	}
	domain := args[1] //获取输入的第一个参数
	log.Println("domain:" + domain)
	// 加载指纹
	sortPairs, webdata := until.ParseCmsDataFromFile("cms.json")

	// 开始并发相关
	t1 := time.Now()
	//domain := "https://www.hacking8.com" // 域名

	newWorker := engine.NewWorker(10)
	newWorker.Checkout(domain)
	newWorker.Start()

	for _, v := range sortPairs {
		tmp_job := engine.JobStruct{domain, v.Path, webdata[v.Path]}
		//fmt.Println(tmp_job)
		newWorker.Add(tmp_job)
	}

	newWorker.Run()

	elapsed := time.Since(t1)
	fmt.Println("elapsed:", elapsed)

}
