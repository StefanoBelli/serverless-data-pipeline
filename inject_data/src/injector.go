package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

var columnNoiseGens = []ColumnNoiseGenerator{
	{
		columnName: "",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return "-1"
			} else {
				return ""
			}
		},
	},
	{
		columnName: "VendorID",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.25 {
				return "0"
			} else {
				return "3"
			}
		},
	},
	{
		columnName: "tpep_pickup_datetime",
		callback: func(value string) string {
			num := rand.Float32()
			if num > 0.8 {
				return ""
			}

			dt, _ := time.Parse("2006-01-02 15:04:05", value)
			if num < 0.7 {
				dt.AddDate(0, 1, 0)
				return dt.Format("2006-01-02 15:04:05")
			} else {
				return dt.Format("Jan 2 '06 at 15:04")
			}
		},
	},
	{
		columnName: "tpep_dropoff_datetime",
		callback: func(value string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			}

			dt, _ := time.Parse("2006-01-02 15:04:05", value)
			return dt.Format("Jan 2 '06 at 15:04")
		},
	},
	{
		columnName: "passenger_count",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.25 {
				return "8.0"
			} else {
				return "0.5"
			}
		},
	},
	{
		columnName: "trip_distance",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "-1.0"
			}
		},
	},
	{
		columnName: "RatecodeID",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.3 {
				return "0.0"
			} else {
				return "7.0"
			}
		},
	},
	{
		columnName: "store_and_fwd_flag",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.3 {
				return "K"
			} else {
				return "19"
			}
		},
	},
	{
		columnName: "PULocationID",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.3 {
				return "-1"
			} else {
				return "pid"
			}
		},
	},
	{
		columnName: "DOLocationID",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.3 {
				return "-1"
			} else {
				return "did"
			}
		},
	},
	{
		columnName: "payment_type",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else if num < 0.3 {
				return "0"
			} else {
				return "7"
			}
		},
	},
	{
		columnName: "fare_amount",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "6.7"
			}
		},
	},
	{
		columnName: "extra",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "?"
			}
		},
	},
	{
		columnName: "mta_tax",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "0.9"
			}
		},
	},
	{
		columnName: "tip_amount",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "0.5"
			}
		},
	},
	{
		columnName: "tolls_amount",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "t"
			}
		},
	},
	{
		columnName: "improvement_surcharge",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "-1.0a"
			}
		},
	},
	{
		columnName: "total_amount",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "0.0"
			}
		},
	},
	{
		columnName: "congestion_surcharge",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "7.4"
			}
		},
	},
	{
		columnName: "Airport_fee",
		callback: func(_ string) string {
			num := rand.Float32()
			if num > 0.5 {
				return ""
			} else {
				return "8.9"
			}
		},
	},
}

var tupleWiseNoiseGens = []TupleWiseNoiseGenerator{
	func(s *string) {
		*s = strings.ReplaceAll(*s, ",", ";")
	},
	func(s *string) {
		*s += ","
	},
	func(s *string) {
		*s = "," + *s
	},
	func(s *string) {
		*s = ""
	},
	func(s *string) {
		i := strings.LastIndexByte(*s, ',')
		*s = (*s)[:i] + (*s)[i+1:]
	},
}

func inject(path string) error {
	var genChans ColumnNoiseGenerationChannels

	genChans.outEntry = make(chan string)
	genChans.outErr = make(chan error)

	go generateColumnNoise(path, columnNoiseGens, genChans)

	fmt.Printf(" --> Injected 0 entries\r")

	i := 1
	for range genChans.outEntry {
		//fmt.Println(entry) //TODO later replace by HTTP GET to endpoint
		fmt.Printf(" --> Injected %d entries\r", i)
		i++
	}

	fmt.Println()

	log.Println("Done")

	myErr := <-genChans.outErr
	close(genChans.outErr)

	return myErr
}

func makeHttpPost(body *string) *string {
	return nil
}
