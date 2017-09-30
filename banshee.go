package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/70data/golang-prometheus/prometheus"
	"github.com/70data/golang-prometheus/prometheus/promhttp"
)

// {"type":"kv", "app":"center","metric":"test","value":"12345", "timeout":"60", "method":"GET", "protocol":"HTTP"}
// {"type":"kv", "app":"center","metric":"test","value":"123456", "timeout":"", "method":"POST", "protocol":"HTTPS"}

var ReqQueue chan string

var (
	LabelStore map[string]int64
	lsSync     sync.Mutex
)

var (
	ValueStore map[string]interface{}
	vsSync     sync.Mutex
)

var (
	TimeOutLabelStore map[string]string
	toLabelSync       sync.Mutex
)

var (
	TimeOutLineStore map[string]int64
	toLineSync       sync.Mutex
)

func mapSort(naiveMap map[string]string) []string {
	sortedKeys := make([]string, 0)
	for indexK, _ := range naiveMap {
		if indexK != "type" && indexK != "metric" && indexK != "value" && indexK != "timeout" {
			sortedKeys = append(sortedKeys, indexK)
		}
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func timeOutMark(cur int64, timeOutMap map[string]string) {
	timeOut := timeOutMap["timeout"]
	if timeOut != "" {
		delete(timeOutMap, "value")
		resaultJSON, err := json.Marshal(timeOutMap)
		if err != nil {
			fmt.Println(err)
		}
		resaultBase := base64.StdEncoding.EncodeToString(resaultJSON)
		toLabelSync.Lock()
		TimeOutLabelStore[resaultBase] = timeOut
		toLabelSync.Unlock()
		toLineSync.Lock()
		TimeOutLineStore[resaultBase] = cur
		toLineSync.Unlock()
	}
}

// init prometheus struct
func dataInit(dataInitMap map[string]string) {
	var customLabels []string
	var valueArray []string
	metric := dataInitMap["metric"]
	value := dataInitMap["value"]
	n, _ := strconv.ParseFloat(value, 64)
	initKeyArray := mapSort(dataInitMap)
	for _, v := range initKeyArray {
		customLabels = append(customLabels, v)
		valueArray = append(valueArray, dataInitMap[v])
	}
	vsSync.Lock()
	ValueStore[metric] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metric,
		Help: "custom info.",
	}, customLabels)
	prometheus.MustRegister(ValueStore[metric].(*prometheus.GaugeVec))
	ValueStore[metric].(*prometheus.GaugeVec).WithLabelValues(valueArray...).Set(n)
	vsSync.Unlock()
	fmt.Println(dataInitMap)
}

// add value to prometheus
func dataConvert(dataMap map[string]string) {
	var valueArray []string
	metric := dataMap["metric"]
	value := dataMap["value"]
	n, _ := strconv.ParseFloat(value, 64)
	convertKeyArray := mapSort(dataMap)
	for _, v := range convertKeyArray {
		valueArray = append(valueArray, dataMap[v])
	}
	vsSync.Lock()
	ValueStore[metric].(*prometheus.GaugeVec).WithLabelValues(valueArray...).Set(n)
	vsSync.Unlock()
}

func Init() {
	LabelStore = make(map[string]int64)
	ValueStore = make(map[string]interface{})
	TimeOutLabelStore = make(map[string]string)
	TimeOutLineStore = make(map[string]int64)
	ReqQueue = make(chan string, 100000)
}

func Process() {
	numCPU := runtime.NumCPU()
	for i := 1; i <= numCPU; i++ {
		go func() {
			for {
				bodyStr := <-ReqQueue
				var resaultMap map[string]string
				json.Unmarshal([]byte(bodyStr), &resaultMap)
				metric := resaultMap["metric"]
				cur := time.Now().Unix()
				lsSync.Lock()
				if _, ok := LabelStore[metric]; ok {
					LabelStore[metric] = cur
					dataConvert(resaultMap)
					timeOutMark(cur, resaultMap)
				} else {
					LabelStore[metric] = cur
					dataInit(resaultMap)
					timeOutMark(cur, resaultMap)
				}
				lsSync.Unlock()
			}
		}()
	}
}

func timeOutMarkDelete() {
	monitorTimeOut := time.NewTicker(60 * time.Second)
	for {
		<-monitorTimeOut.C
		nowTime := time.Now().Unix()
		toLabelSync.Lock()
		toLineSync.Lock()
		for resaultBase, timeLineStr := range TimeOutLabelStore {
			resaultBytes, _ := base64.StdEncoding.DecodeString(resaultBase)
			timeLine, _ := strconv.ParseInt(timeLineStr, 10, 64)
			lastMarkTime := TimeOutLineStore[resaultBase]
			// delete time out data
			if timeLine < (nowTime - lastMarkTime) {
				var metricInfoTemp map[string]string
				var valueArray []string
				json.Unmarshal(resaultBytes, &metricInfoTemp)
				metric := metricInfoTemp["metric"]
				deleteKeys := make([]string, 0)
				for k, _ := range metricInfoTemp {
					if k != "type" && k != "metric" && k != "timeout" {
						deleteKeys = append(deleteKeys, k)
					}
				}
				for _, v := range deleteKeys {
					valueArray = append(valueArray, metricInfoTemp[v])
				}
				sort.Strings(deleteKeys)
				vsSync.Lock()
				ValueStore[metric].(*prometheus.GaugeVec).DeleteLabelValues(valueArray...)
				vsSync.Unlock()
				delete(TimeOutLabelStore, resaultBase)
				delete(TimeOutLineStore, resaultBase)
				fmt.Println(metricInfoTemp)
			}
		}
		toLabelSync.Unlock()
		toLineSync.Unlock()
	}
}

// Receive custom data.
func customData(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	res.Write([]byte("succeed"))
	ReqQueue <- string(body)
	req.Body.Close()
}

func main() {
	Init()
	go Process()
	go timeOutMarkDelete()
	http.HandleFunc("/customData/", customData)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2336", nil)
}
