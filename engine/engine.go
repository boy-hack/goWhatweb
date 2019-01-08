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
	"sync"
	"time"
)

type Worker struct {
	MaxPool     int
	MaxQueue    int
	JobQueue    chan JobStruct
	quit        chan bool
	ResultChain chan until.Singcms
	domain      string // 域名
	delay       int    // 访问的延时时间
	wg          *sync.WaitGroup
	finished    bool // 是否完成
	mutex       sync.Mutex
}

type JobStruct struct {
	Domain  string
	Path    string
	Cmsdata []until.Singcms
}

func NewWorker(count int, domain string, wg *sync.WaitGroup) Worker {
	return Worker{
		MaxPool:     count,
		quit:        make(chan bool, count),
		JobQueue:    make(chan JobStruct, count),
		ResultChain: make(chan until.Singcms),
		domain:      domain,
		delay:       0,
		wg:          wg,
		finished:    false,
	}
}

func (w *Worker) Start() {
	// starting n number of workers
	for i := 0; i < w.MaxPool; i++ {
		go func(i int) {
			for {
				select {
				case j := <-w.JobQueue:
					Comsumer(j, w)
				}
			}
		}(i)
	}
	go w.Run()
}

func (w *Worker) Checkout() {
	bytes, headers, err := fetch.Get(w.domain)
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
		log.Printf("domain:%s waf:%s", w.domain, wafname)
		w.MaxPool = 1
		w.delay = 200
	}

}

func (w *Worker) Stop() {
	w.mutex.Lock()
	w.finished = true
	w.mutex.Unlock()
}

func (w *Worker) Add(i JobStruct) {
	//fmt.Println(i)
	w.mutex.Lock()
	status := w.finished
	w.mutex.Unlock()
	if status {
		return
	}
	w.wg.Add(1)
	w.JobQueue <- i
}

func (w *Worker) Run() {
	time.Sleep(time.Second)
	for {
		r := <-w.ResultChain
		stdout := "{Domain:%s Cms:%s Path:%s Option:%s Content:%s}"
		log.Printf(stdout, w.domain, r.Name, r.Path, r.Option, r.Content)
		break

		//select {
		//case r := <-w.ResultChain:
		//default:
		//	if len(w.JobQueue) == 0 {
		//		return
		//	}
		//	time.Sleep(time.Millisecond * 20)
		//}
	}
}

func Comsumer(job JobStruct, w *Worker) {
	url := job.Domain + job.Path
	resp, e := fetch.Head(url)
	if e != nil {
		log.Println(e)
		defer w.wg.Done()
		return
	}
	if resp.StatusCode != 200 {
		defer w.wg.Done()
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
				break
			}
		} else if option == "md5" {
			md5str := fmt.Sprintf("%x", md5.Sum(content))
			if md5str == cmsinfo.Content {
				//fmt.Println(cmsinfo)
				w.ResultChain <- cmsinfo
				w.Stop()
				break
			}

		}
	}
	defer w.wg.Done()
}
