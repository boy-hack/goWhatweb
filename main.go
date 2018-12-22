package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type singcms struct {
	Path    string
	Option  string
	Content string
	Name    string
}

type Pair struct {
	Path   string
	length int
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].length > p[j].length }

func sortMapByValue(m map[string]int) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(p)
	return p
}

func main() {

	s, err := ioutil.ReadFile("cms.json")
	if err != nil {
		panic(err)
	}
	var f map[string][]singcms
	err = json.Unmarshal(s, &f)

	webdata := make(map[string][]singcms) // 实际数据

	for k, v := range f {
		//_,ok := webdata[k]
		//
		//if(!ok){
		//	webdata[k] = make([]singcms,len(v))
		//}
		//
		for _, data := range v {
			tmp_path := data.Path
			_, ok := webdata[tmp_path]
			if !ok {
				webdata[tmp_path] = make([]singcms, 0)
			}
			data.Name = k
			webdata[tmp_path] = append(webdata[tmp_path], data)
		}
	}
	sortdata := make(map[string]int) // 排序数据
	for k, v := range webdata {
		sortdata[k] = len(v)
	}
	sortPairs := sortMapByValue(sortdata)
	//fmt.Println(len(sortMapByValue(sortdata)))
	// 开始并发相关
	domain := "https://www.t00ls.net" // 域名
	count := make(chan int, 200)      // 并发总量
	//quite := make(chan int) // 控制退出

	Client := http.Client{
		Timeout: 30 * time.Second,
	}
	for _, v := range sortPairs {
		path2 := v.Path
		domain_url := domain + path2
		count <- 1
		go func() {
			//fmt.Println(domain_url)
			resp, err := Client.Head(domain_url)
			if err != nil {
				log.Println(err)
				<-count
				return
			}
			//fmt.Println(resp.StatusCode)
			//tmp := "/uc_server/view/default/admin_frame_main.htm"
			//var debug bool
			//debug = false
			//if(strings.Contains(domain_url,tmp)){
			//	debug = true
			//}
			//if(debug){
			//	fmt.Println(resp.StatusCode,domain_url)
			//}
			if resp.StatusCode == 404 {
				<-count
				return
			}
			response, err := Client.Get(domain_url)
			if err != nil {
				log.Println(err)
				<-count
				return
			}
			defer response.Body.Close()

			bytes, _ := ioutil.ReadAll(response.Body)
			//if(debug){
			//	fmt.Println(string(bytes))
			//}
			cmsinfos := webdata[path2]
			for _, cmsinfo := range cmsinfos {
				option := cmsinfo.Option
				if option == "keyword" {
					if strings.Contains(string(bytes), cmsinfo.Content) {
						fmt.Println(cmsinfo)
						<-count
						return
					}
				} else if option == "md5" {
					md5str := fmt.Sprintf("%x", md5.Sum(bytes))
					if md5str == cmsinfo.Content {
						fmt.Println(cmsinfo)
						<-count
						return
					}

				}
			}
			<-count
			return
		}()

	}
	for {
		if len(count) == 0 {
			fmt.Println("over")
			break
		}
		time.Sleep(time.Second)
	}
}
