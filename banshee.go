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

var LabelStore map[string]int64
var ValueStore map[string]interface{}

func Init() {
	LabelStore = make(map[string]int64)
	ValueStore = make(map[string]interface{})
}

// init prometheus struct
func dataInit(resaultMap map[string]string) {
	var customLabels []string
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
		customLabels = append(customLabels, v)
		valueArray = append(valueArray, resaultMap[v])
	}
	ValueStore[metric] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metric,
		Help: "custom info.",
	}, customLabels)
	prometheus.MustRegister(ValueStore[metric].(*prometheus.GaugeVec))
	fmt.Println(valueArray)
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
	fmt.Println(valueArray)
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
	if _, ok := LabelStore[metric]; ok {
		go dataConvert(resaultMap)

	} else {
		cur := time.Now().Unix()
		LabelStore[metric] = cur
		fmt.Println(LabelStore)
		dataInit(resaultMap)
	}
	res.Write([]byte("succeed"))
}

func main() {
	Init()
	http.HandleFunc("/customData/", customData)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2336", nil)
}
