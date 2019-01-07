package engine

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"goWhatweb/fetch"
	"goWhatweb/until"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type Worker struct {
	MaxPool     int
	MaxQueue    int
	JobQueue    chan JobStruct
	quit        chan bool
	ResultChain chan until.Singcms
}

type JobStruct struct {
	Domain  string
	Path    string
	Cmsdata []until.Singcms
}

func NewWorker(count int) Worker {
	return Worker{
		MaxPool:     count,
		quit:        make(chan bool, count),
		JobQueue:    make(chan JobStruct, count),
		ResultChain: make(chan until.Singcms),
	}
}

func (w *Worker) Start() {
	// starting n number of workers
	for i := 0; i < w.MaxPool; i++ {
		go func(i int) {
			//fmt.Println("w", i)
			for {
				select {
				case <-w.quit:
					//fmt.Printf("%d channel quit\n",i)
					return
				case j := <-w.JobQueue:
					Comsumer(j, w)

				}

			}
		}(i)
	}
}

func (w *Worker) Checkout(domain string) {
	bytes, headers, err := fetch.Get(domain)
	if err != nil {
		panic(err)
	}
	// waf识别
	fi, err := os.Open("waf.txt")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	wafname := ""
	for {
		s, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		str := string(s)
		strs := strings.Split(str, "|")
		option := strs[1]
		content := strs[3]
		if option == "index" {
			if strings.Contains(string(bytes), content) {
				wafname = strs[0]
				break
			}
		} else {
			val, ok := headers[strs[2]]
			if ok {
				match, _ := regexp.MatchString(content, strings.Join(val, ""))
				if match {
					wafname = strs[0]
					break
				}
			}
		}
	}
	if wafname != "" {
		fmt.Printf("domain:%s waf:%s", domain, wafname)
		w.MaxPool = 5
	}

}

func (w *Worker) Stop() {
	for i := 0; i < w.MaxPool; i++ {
		w.quit <- true
	}
}

func (w *Worker) Add(i JobStruct) {
	//fmt.Println(i)
	go func() {
		w.JobQueue <- i
	}()
}

func (w *Worker) Run() {
	time.Sleep(time.Second)
	for {
		select {
		case r := <-w.ResultChain:
			stdout := "{Cms:%s Path:%s Option:%s Content:%s}"
			log.Printf(stdout, r.Name, r.Path, r.Option, r.Content)
			return
		default:
			if len(w.JobQueue) == 0 {
				return
			}
			time.Sleep(time.Millisecond * 20)
		}

	}
}

func Comsumer(job JobStruct, w *Worker) {
	url := job.Domain + job.Path
	//fmt.Println(url)
	resp, e := fetch.Head(url)
	if e != nil {
		log.Println(e)
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	content, _, _ := fetch.Get(url)

	cmsinfos := job.Cmsdata
	for _, cmsinfo := range cmsinfos {
		option := cmsinfo.Option
		if option == "keyword" {
			if strings.Contains(string(content), cmsinfo.Content) {
				//fmt.Println(cmsinfo)
				w.ResultChain <- cmsinfo
				w.Stop()
				return
			}
		} else if option == "md5" {
			md5str := fmt.Sprintf("%x", md5.Sum(content))
			if md5str == cmsinfo.Content {
				//fmt.Println(cmsinfo)
				w.ResultChain <- cmsinfo
				w.Stop()
				return
			}

		}
	}
}
