package engine

import (
	"crypto/md5"
	"fmt"
	"goWhatweb/fetch"
	"goWhatweb/until"
	"log"
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
	content, _ := fetch.Get(url)

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
