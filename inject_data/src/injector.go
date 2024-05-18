package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
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

type RequestBody struct {
	Tuple string `json:"tuple"`
}

type ResponseBody struct {
	ExecutionArn string `json:"executionArn"`
	//StartDate    uint64 `json:"startDate"`
}

func inject(path string) error {
	var genChans ColumnNoiseGenerationChannels

	genChans.outEntry = make(chan string)
	genChans.outErr = make(chan error)

	go readTuplesAndGenerateColumnNoise(path, columnNoiseGens, genChans)

	fmt.Printf(" --> Injected 0 entries\r")

	i := 1
	for entry := range genChans.outEntry {
		generateTupleWiseNoise(&entry, &tupleWiseNoiseGens)
		reqBodyBytes, err := json.Marshal(&RequestBody{Tuple: entry})
		if err != nil {
			log.Printf(" --> Unable to parse JSON (ignoring): %s\n",
				err.Error())
		} else {
			resBodyBytes, err := makeHttpPost(&reqBodyBytes)
			if err != nil {
				log.Printf(" --> HTTP client error (ignoring): %s - %s\n",
					string(resBodyBytes), err.Error())
			} else {
				smExec := ResponseBody{}
				if err := json.Unmarshal(resBodyBytes, &smExec); err != nil {
					log.Printf(" --> Unable to parse JSON (ignoring): %s\n", err)
				} else {
					fmt.Printf(" --> Injected %d entries. exec = { arn: %s, start: [...] }\r",
						i, smExec.ExecutionArn)
					i++
				}
			}
		}
	}

	fmt.Println()

	log.Println("Done")

	myErr := <-genChans.outErr
	close(genChans.outErr)

	return myErr
}

func makeHttpPost(body *[]byte) ([]byte, error) {
	url := programConfig.injector.http.apiEndpoint + "/store"

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(*body))
	if err != nil {
		return []byte("request builder"), err
	}

	req.Header.Add("Content-Type", "application/json")

	if len(programConfig.injector.http.authKey) != 0 {
		req.Header.Add("Authorization", programConfig.injector.http.authKey)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return []byte("request issuer (http client)"), err
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte("response body buffer reader"), err
	}

	return resBody, nil
}
