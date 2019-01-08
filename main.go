package main

import (
	"fmt"
	"goWhatweb/engine"
	"goWhatweb/until"
	"log"
	"sync"
	"time"
)

func main() {
	//args := os.Args //获取用户输入的所有参数
	//
	//if args == nil || len(args) != 2 {
	//	log.Fatal("err:./goWhatweb https://x.hacking8.com")
	//}
	//domain := args[1] //获取输入的第一个参数
	//log.Println("domain:" + domain)
	domains := []string{"https://www.hacking8.com", "https://x.hacking8.com"}

	// 加载指纹
	sortPairs, webdata := until.ParseCmsDataFromFile("cms.json")
	var wg sync.WaitGroup

	// 开始并发相关
	t1 := time.Now()
	fmt.Println("Load url:", domains)
	for _, domain := range domains {
		go func(d string) {
			newWorker := engine.NewWorker(7, d, &wg)
			newWorker.Checkout()
			newWorker.Start()
			for _, v := range sortPairs {
				tmp_job := engine.JobStruct{d, v.Path, webdata[v.Path]}
				//fmt.Println(tmp_job)
				newWorker.Add(tmp_job)
			}
		}(domain)
	}
	time.Sleep(time.Second * 2)
	log.Println("初始化完成")

	wg.Wait()
	elapsed := time.Since(t1)
	fmt.Println("elapsed:", elapsed)

}
