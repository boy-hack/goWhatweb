package until

import (
	"encoding/json"
	"io/ioutil"
	"sort"
)

type Singcms struct {
	Path    string
	Option  string
	Content string
	Name    string
}

func ParseCmsDataFromFile(filename string) (PairList, map[string][]Singcms) {
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var f map[string][]Singcms
	err = json.Unmarshal(s, &f)

	webdata := make(map[string][]Singcms) // 实际数据

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
				webdata[tmp_path] = make([]Singcms, 0)
			}
			data.Name = k
			webdata[tmp_path] = append(webdata[tmp_path], data)
		}
	}
	sortdata := make(map[string]int)
	for k, v := range webdata {
		sortdata[k] = len(v)
	}
	sortPairs := sortMapByValue(sortdata) // 排序数据
	return sortPairs, webdata

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
