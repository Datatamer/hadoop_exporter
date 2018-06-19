package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "datanode"
)

var (
	listenAddress  = flag.String("web.listen-address", ":9072", "Address on which to expose metrics and web interface.")
	metricsPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	datanodeJmxUrl = flag.String("datanode.jmx.url", "http://localhost:50075/jmx", "Hadoop JMX URL.")
)

type Exporter struct {
	url               string
	CapacityTotal     prometheus.Gauge
	CapacityUsed      prometheus.Gauge
	CapacityRemaining prometheus.Gauge

	CacheCapacity prometheus.Gauge
	CacheUsed     prometheus.Gauge

	FailedVolumes         prometheus.Gauge
	EstimatedCapacityLost prometheus.Gauge

	BlocksCached          prometheus.Gauge
	BlocksFailedToCache   prometheus.Gauge
	BlocksFailedToUncache prometheus.Gauge

	heapMemoryUsageCommitted prometheus.Gauge
	heapMemoryUsageInit      prometheus.Gauge
	heapMemoryUsageMax       prometheus.Gauge
	heapMemoryUsageUsed      prometheus.Gauge
}

func NewExporter(url string) *Exporter {
	return &Exporter{
		url: url,
		CapacityTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "CapacityTotal",
			Help:      "CapacityTotal",
		}),
		CapacityUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "CapacityUsed",
			Help:      "CapacityUsed",
		}),
		CapacityRemaining: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "CapacityRemaining",
			Help:      "CapacityRemaining",
		}),
		CacheCapacity: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "CacheCapacity",
			Help:      "CacheCapacity",
		}),
		CacheUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "CacheUsed",
			Help:      "CacheUsed",
		}),

		FailedVolumes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "FailedVolumes",
			Help:      "FailedVolumes",
		}),
		EstimatedCapacityLost: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "EstimatedCapacityLost",
			Help:      "EstimatedCapacityLost",
		}),

		BlocksCached: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "BlocksCached",
			Help:      "BlocksCached",
		}),
		BlocksFailedToCache: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "BlocksFailedToCache",
			Help:      "BlocksFailedToCache",
		}),
		BlocksFailedToUncache: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "BlocksFailedToUncache",
			Help:      "BlocksFailedToUncache",
		}),

		heapMemoryUsageCommitted: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageCommitted",
			Help:      "heapMemoryUsageCommitted",
		}),
		heapMemoryUsageInit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageInit",
			Help:      "heapMemoryUsageInit",
		}),
		heapMemoryUsageMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageMax",
			Help:      "heapMemoryUsageMax",
		}),
		heapMemoryUsageUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageUsed",
			Help:      "heapMemoryUsageUsed",
		}),
	}
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.CapacityTotal.Describe(ch)
	e.CapacityUsed.Describe(ch)
	e.CapacityRemaining.Describe(ch)
	e.CacheCapacity.Describe(ch)
	e.CacheUsed.Describe(ch)
	e.FailedVolumes.Describe(ch)
	e.EstimatedCapacityLost.Describe(ch)
	e.BlocksCached.Describe(ch)
	e.BlocksFailedToCache.Describe(ch)
	e.BlocksFailedToUncache.Describe(ch)
	e.heapMemoryUsageCommitted.Describe(ch)
	e.heapMemoryUsageInit.Describe(ch)
	e.heapMemoryUsageMax.Describe(ch)
	e.heapMemoryUsageUsed.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	resp, err := http.Get(e.url)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}
	var f interface{}
	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Error(err)
	}
	if len(data) == 0 {
	  log.Error("Empty.")
	  return
	}
	// {"beans":[{"name":"Hadoop:service=NameNode,name=FSNamesystem", ...}, {"name":"java.lang:type=MemoryPool,name=Code Cache", ...}, ...]}
	m := f.(map[string]interface{})
	// [{"name":"Hadoop:service=NameNode,name=FSNamesystem", ...}, {"name":"java.lang:type=MemoryPool,name=Code Cache", ...}, ...]
	var nameList = m["beans"].([]interface{})
	for _, nameData := range nameList {
		nameDataMap := nameData.(map[string]interface{})
		/*
			{
				"name" : "Hadoop:service=DataNode,name=FSDatasetState",
				"modelerType" : "org.apache.hadoop.hdfs.server.datanode.fsdataset.impl.FsDatasetImpl",
				"Remaining" : 49909760000,
				"StorageInfo" : "FSDataset{dirpath='[/tmp/hadoop-root/dfs/data/current]'}",
				"Capacity" : 228769484800,
				"DfsUsed" : 327680,
				"CacheCapacity" : 0,
				"CacheUsed" : 0,
				"NumFailedVolumes" : 0,
				"FailedStorageLocations" : [ ],
				"LastVolumeFailureDate" : 0,
				"EstimatedCapacityLostTotal" : 0,
				"NumBlocksCached" : 0,
				"NumBlocksFailedToCache" : 0,
				"NumBlocksFailedToUnCache" : 0
			}
		*/
		if nameDataMap["name"] == "Hadoop:service=DataNode,name=FSDatasetState" {
			e.CapacityTotal.Set(nameDataMap["Capacity"].(float64))
			e.CapacityUsed.Set(nameDataMap["DfsUsed"].(float64))
			e.CapacityRemaining.Set(nameDataMap["Remaining"].(float64))

			e.CacheCapacity.Set(nameDataMap["CacheCapacity"].(float64))
			e.CacheUsed.Set(nameDataMap["CacheUsed"].(float64))

			e.FailedVolumes.Set(nameDataMap["NumFailedVolumes"].(float64))
			e.EstimatedCapacityLost.Set(nameDataMap["EstimatedCapacityLostTotal"].(float64))

			e.BlocksCached.Set(nameDataMap["NumBlocksCached"].(float64))
			e.BlocksFailedToCache.Set(nameDataMap["NumBlocksFailedToCache"].(float64))
			e.BlocksFailedToUncache.Set(nameDataMap["NumBlocksFailedToUnCache"].(float64))
		}
		/*
			   {
				"name" : "java.lang:type=Memory",
				"modelerType" : "sun.management.MemoryImpl",
				"Verbose" : false,
				"HeapMemoryUsage" : {
					"committed" : 312999936,
					"init" : 326803392,
					"max" : 932184064,
					"used" : 50282512
				},
					"NonHeapMemoryUsage" : {
					"committed" : 30343168,
					"init" : 24576000,
					"max" : 136314880,
					"used" : 29086488
				},
					"ObjectPendingFinalizationCount" : 0,
					"ObjectName" : "java.lang:type=Memory"
				}
		*/
		if nameDataMap["name"] == "java.lang:type=Memory" {
			heapMemoryUsage := nameDataMap["HeapMemoryUsage"].(map[string]interface{})
			e.heapMemoryUsageCommitted.Set(heapMemoryUsage["committed"].(float64))
			e.heapMemoryUsageInit.Set(heapMemoryUsage["init"].(float64))
			e.heapMemoryUsageMax.Set(heapMemoryUsage["max"].(float64))
			e.heapMemoryUsageUsed.Set(heapMemoryUsage["used"].(float64))
		}
	}
	e.CapacityTotal.Collect(ch)
	e.CapacityUsed.Collect(ch)
	e.CapacityRemaining.Collect(ch)
	e.CacheCapacity.Collect(ch)
	e.CacheUsed.Collect(ch)
	e.FailedVolumes.Collect(ch)
	e.EstimatedCapacityLost.Collect(ch)
	e.BlocksCached.Collect(ch)
	e.BlocksFailedToCache.Collect(ch)
	e.BlocksFailedToUncache.Collect(ch)
	e.heapMemoryUsageCommitted.Collect(ch)
	e.heapMemoryUsageInit.Collect(ch)
	e.heapMemoryUsageMax.Collect(ch)
	e.heapMemoryUsageUsed.Collect(ch)
}

func main() {
	flag.Parse()

	exporter := NewExporter(*datanodeJmxUrl)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
        <head><title>DataNode Exporter</title></head>
        <body>
        <h1>DataNode Exporter</h1>
        <p><a href="` + *metricsPath + `">Metrics</a></p>
        </body>
        </html>`))
	})
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
