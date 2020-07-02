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

package influx

import (
	"errors"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/tests/services"
	influxLib "github.com/orourkedd/influxdb1-client"
	"github.com/orourkedd/influxdb1-client/models"
	"testing"
)

func TestUtil(t *testing.T) {
	influxClientMock := services.NewClientMock()

	influxClient := Influx{
		config: &configuration.ConfigStruct{
			Debug: true,
		},
		client: &influxClientMock,
	}

	t.Run("executeQuery", func(t *testing.T) {
		t.Run("net error", func(t *testing.T) {
			influxClientMock.SetQueryResponse(nil, netError{
				error: errors.New("net error"),
			})
			_, err := influxClient.executeQuery("test", "test")
			if err != ErrInfluxConnection {
				t.Fail()
			}
		})
		t.Run("other err", func(t *testing.T) {
			testErr := errors.New("other err")
			influxClientMock.SetQueryResponse(nil, testErr)
			_, err := influxClient.executeQuery("test", "test")
			if err != testErr {
				t.Fail()
			}
		})
		t.Run("response nil", func(t *testing.T) {
			influxClientMock.SetQueryResponse(nil, nil)
			_, err := influxClient.executeQuery("test", "test")
			if err != ErrNULL {
				t.Fail()
			}
		})
		t.Run("response not found", func(t *testing.T) {
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Err: errors.New("DB test not found"),
			}, nil)
			_, err := influxClient.executeQuery("test", "test")
			if err != ErrNotFound {
				t.Fail()
			}
		})
		t.Run("response normal", func(t *testing.T) {
			expect := &influxLib.Response{
				Results: []influxLib.Result{
					{
						Series: []models.Row{
							{
								Name: "test",
							},
						},
					},
				},
			}
			influxClientMock.SetQueryResponse(expect, nil)
			actual, err := influxClient.executeQuery("test", "test")
			if err != nil {
				t.Fail()
			}
			if actual != expect {
				t.Fail()
			}
		})
	})
}

type netError struct {
	error
}

func (n netError) Error() string {
	return n.error.Error()
}

func (n netError) Timeout() bool {
	return true
}
func (n netError) Temporary() bool {
	return true
}