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

// {"type":"kv", "app":"sre","metric":"gatewaytest","value":"12345", "timeout":"60", "method":"GET", "protocol":"HTTP"}
// {"type":"kv", "app":"sre","metric":"gatewaytest","value":"12345", "timeout":"", "method":"POST", "protocol":"HTTP"}

var (
	LabelStore        map[string]int64
	ValueStore        map[string]interface{}
	TimeOutLabelStore map[string]string
	TimeOutLineStore  map[string]int64
)

func Init() {
	LabelStore = make(map[string]int64)
	ValueStore = make(map[string]interface{})
	TimeOutLabelStore = make(map[string]string)
	TimeOutLineStore = make(map[string]int64)
}

func timeOutMark(cur int64, resaultMap map[string]string) {
	timeout := resaultMap["timeout"]
	if timeout != "" {
		delete(resaultMap, "value")
		resaultJSON, _ := json.Marshal(resaultMap)
		resaultBase := base64.StdEncoding.EncodeToString(resaultJSON)
		TimeOutLabelStore[resaultBase] = timeout
		TimeOutLineStore[resaultBase] = cur
	}
}

func timeOutMarkDelete() {
	monitorTimeOut := time.NewTicker(60 * time.Second)
	for {
		<-monitorTimeOut.C
		nowTime := time.Now().Unix()
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
				sortedKeys := make([]string, 0)
				for indexK, _ := range metricInfoTemp {
					if indexK != "type" && indexK != "metric" && indexK != "timeout" {
						sortedKeys = append(sortedKeys, indexK)
					}
				}
				for _, v := range sortedKeys {
					valueArray = append(valueArray, metricInfoTemp[v])
				}
				sort.Strings(sortedKeys)
				fmt.Println(sortedKeys, valueArray)
				ValueStore[metric].(*prometheus.GaugeVec).DeleteLabelValues(valueArray...)
				delete(TimeOutLabelStore, resaultBase)
				delete(TimeOutLineStore, resaultBase)
			}
		}
	}
}

// init prometheus struct
func dataInit(resaultMap map[string]string) {
	var customLabels []string
	var valueArray []string
	metric := resaultMap["metric"]
	value := resaultMap["value"]
	n, _ := strconv.ParseFloat(value, 64)
	sortedKeys := make([]string, 0)
	for indexK, _ := range resaultMap {
		if indexK != "type" && indexK != "metric" && indexK != "value" && indexK != "timeout" {
			sortedKeys = append(sortedKeys, indexK)
		}
	}
	sort.Strings(sortedKeys)
	for _, v := range sortedKeys {
		customLabels = append(customLabels, v)
		valueArray = append(valueArray, resaultMap[v])
	}
	ValueStore[metric] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metric,
		Help: "custom info.",
	}, customLabels)
	prometheus.MustRegister(ValueStore[metric].(*prometheus.GaugeVec))
	ValueStore[metric].(*prometheus.GaugeVec).WithLabelValues(valueArray...).Set(n)
}

// add value to prometheus
func dataConvert(resaultMap map[string]string) {
	var valueArray []string
	metric := resaultMap["metric"]
	value := resaultMap["value"]
	n, _ := strconv.ParseFloat(value, 64)
	sorted_keys := make([]string, 0)
	for indexK, _ := range resaultMap {
		if indexK != "type" && indexK != "metric" && indexK != "value" && indexK != "timeout" {
			sorted_keys = append(sorted_keys, indexK)
		}
	}
	sort.Strings(sorted_keys)
	for _, v := range sorted_keys {
		valueArray = append(valueArray, resaultMap[v])
	}
	ValueStore[metric].(*prometheus.GaugeVec).WithLabelValues(valueArray...).Set(n)
}

// Receive custom data.
func customData(res http.ResponseWriter, req *http.Request) {
	len := req.ContentLength
	body := make([]byte, len)
	defer req.Body.Close()
	req.Body.Read(body)
	var resaultMap map[string]string
	json.Unmarshal([]byte(string(body)), &resaultMap)
	metric := resaultMap["metric"]
	cur := time.Now().Unix()
	if _, ok := LabelStore[metric]; ok {
		LabelStore[metric] = cur
		go dataConvert(resaultMap)
		go timeOutMark(cur, resaultMap)
	} else {
		LabelStore[metric] = cur
		fmt.Println(LabelStore)
		dataInit(resaultMap)
		go timeOutMark(cur, resaultMap)
	}
	res.Write([]byte("succeed"))
}

func main() {
	Init()
	go timeOutMarkDelete()
	http.HandleFunc("/customData/", customData)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2336", nil)
}
