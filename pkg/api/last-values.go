/*
 *    Copyright 2020 InfAI (CC SES)
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package api

import (
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	influxdb "github.com/SENERGY-Platform/influx-wrapper/pkg/influx"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func init() {
	endpoints = append(endpoints, LastValuesEndpoint)
}

const userHeader = "X-UserID"

type RequestElement struct {
	Measurement string `json:"measurement"`
	ColumnName  string `json:"columnName"`
}

type ResponseElement struct {
	Time  string      `json:"time"`
	Value interface{} `json:"value"`
}

func LastValuesEndpoint(router *httprouter.Router, config configuration.Config, influx *influxdb.Influx) {
	router.POST("/last-values", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		db := request.Header.Get(userHeader)
		if db == "" {
			http.Error(writer, "Missing header "+userHeader, http.StatusBadRequest)
			return
		}

		var requestElements []RequestElement
		err := json.NewDecoder(request.Body).Decode(&requestElements)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if config.Debug {
			log.Print("Request:")
			err = json.NewEncoder(log.Writer()).Encode(requestElements)
			if err != nil {
				log.Println(err)
			}
		}

		responseElements := []ResponseElement{}

		for _, requestElement := range requestElements {
			time, value, err := influx.GetLatestValue(db, requestElement.Measurement, requestElement.ColumnName)
			if err != nil {
				switch err {
				case influxdb.ErrInfluxConnection:
					http.Error(writer, err.Error(), http.StatusBadGateway)
					return
				case influxdb.ErrNotFound:
					http.Error(writer, err.Error(), http.StatusNotFound)
					return
				case influxdb.ErrNULL:
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				case influxdb.ErrUnexpectedLength:
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				default:
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			responseElements = append(responseElements, ResponseElement{
				Time:  time,
				Value: value,
			})

		}

		writer.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(writer).Encode(responseElements)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
		}
	})

}