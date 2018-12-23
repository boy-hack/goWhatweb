package main

import (
	"fmt"
	"goWhatweb/engine"
	"goWhatweb/until"
	"time"
)

func main() {
	sortPairs, webdata := until.ParseCmsDataFromFile("cms.json")

	// 开始并发相关
	t1 := time.Now()
	domain := "https://www.t00ls.net" // 域名

	newWorker := engine.NewWorker(10)
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
