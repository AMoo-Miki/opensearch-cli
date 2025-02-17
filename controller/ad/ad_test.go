/*
 * SPDX-License-Identifier: Apache-2.0
 *
 * The OpenSearch Contributors require contributions made to
 * this file be licensed under the Apache-2.0 license or a
 * compatible open source license.
 *
 * Modifications Copyright OpenSearch Contributors. See
 * GitHub history for details.
 */
/*
 * Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package ad

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	mockController "opensearch-cli/controller/platform/mocks"
	entity "opensearch-cli/entity/ad"
	gateway "opensearch-cli/gateway/ad/mocks"
	"opensearch-cli/mapper"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const mockDetectorID = "m4ccEnIBTXsGi3mvMt9p"
const mockDetectorName = "detector"

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
func helperConvertToInterface(input []string) []interface{} {
	s := make([]interface{}, len(input))
	for i, v := range input {
		s[i] = v
	}
	return s
}

func getRawFilter() []byte {
	return []byte(`{
    			"bool": {
      				"filter": [{
          				"exists": {
						"field": "value",
            			"boost": 1
         			 	}
        			}],
      				"adjust_pure_negative": true,
      				"boost": 1
    			}
  			}`)
}
func getFinalFilter(additionalFilters ...json.RawMessage) []byte {

	filter1 := []byte(`{
    			"bool": {
      				"filter": {
          				"term": {
							"ip" : "localhost"
         			 	}
        			}
				}
  			}`)

	if len(additionalFilters) < 1 {
		return filter1
	}
	filter := entity.Query{
		Bool: entity.Bool{
			Must: []json.RawMessage{
				filter1,
			},
		},
	}
	filter.Bool.Must = append(filter.Bool.Must, additionalFilters...)
	marshal, _ := json.Marshal(filter)
	return marshal
}

func getCreateDetectorRequest() entity.CreateDetectorRequest {
	return entity.CreateDetectorRequest{
		Name:        "testdata-detector",
		Description: "Test detector",
		TimeField:   "timestamp",
		Index:       []string{"order*"},
		Features: []entity.FeatureRequest{{
			AggregationType: []string{"sum"},
			Enabled:         true,
			Field:           []string{"value"},
		}},
		Filter:         getRawFilter(),
		Interval:       "1m",
		Delay:          "1m",
		Start:          true,
		PartitionField: mapper.StringToStringPtr("ip"),
	}
}
func getRawFeatureAggregation() []byte {
	return []byte(`{
        			"sum_value": {
          				"sum": {
            				"field": "value"
						}
        			}
      			}`)
}
func getCreateDetector() *entity.CreateDetector {
	return &entity.CreateDetector{
		Name:        "testdata-detector",
		Description: "Test detector",
		TimeField:   "timestamp",
		Index:       []string{"order*"},
		Features: []entity.Feature{
			{
				Name:             "sum_value",
				Enabled:          true,
				AggregationQuery: getRawFeatureAggregation(),
			},
		},
		Filter: getRawFilter(),
		Interval: entity.Interval{
			Period: entity.Period{
				Duration: 1,
				Unit:     "Minutes",
			},
		},
		Delay: entity.Interval{
			Period: entity.Period{
				Duration: 1,
				Unit:     "Minutes",
			},
		},
	}
}
func TestController_StartDetector(t *testing.T) {
	t.Run("start empty detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctx := context.Background()
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		assert.Error(t, ctrl.StartDetector(ctx, ""))
	})
	t.Run("start detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StartDetector(ctx, "detectorID").Return(errors.New("no connection"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		assert.Error(t, ctrl.StartDetector(ctx, "detectorID"))
	})
	t.Run("start detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StartDetector(ctx, "detectorID").Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		assert.NoError(t, ctrl.StartDetector(ctx, "detectorID"))
	})
}

func TestController_StopDetector(t *testing.T) {
	t.Run("stop empty detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctx := context.Background()
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.StopDetector(ctx, "")
		assert.Error(t, err)
	})
	t.Run("stop detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(nil, errors.New("gateway failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.StopDetector(ctx, "detectorID")
		assert.Error(t, err)
	})
	t.Run("stop detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(mapper.StringToStringPtr("Stopped Detector"), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.StopDetector(ctx, "detectorID")
		assert.NoError(t, err)
	})
}

func TestController_CreateAnomalyDetector(t *testing.T) {
	t.Run("gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().CreateDetector(ctx, getCreateDetector()).Return(nil, errors.New("failed to connect"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateAnomalyDetector(ctx, r)
		assert.EqualError(t, err, "failed to connect")
	})
	t.Run("entity failed to create", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().CreateDetector(ctx, getCreateDetector()).Return(nil, errors.New(string(helperLoadBytes(t, "create_failed_response.json"))))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateAnomalyDetector(ctx, r)
		assert.EqualError(t, err, "Cannot create anomaly detector with name [testdata-detector] as it's already used by detector [wR_1XXMBs3q1IVz33Sk-]")
	})
	t.Run("entity succeeded without starting", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		r.Start = false
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().CreateDetector(ctx, getCreateDetector()).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		detectorID, err := ctrl.CreateAnomalyDetector(ctx, r)
		assert.NoError(t, err)
		assert.NotNil(t, detectorID)
		assert.EqualValues(t, mockDetectorID, *detectorID)
	})
	t.Run("entity succeeded", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().CreateDetector(ctx, getCreateDetector()).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockADGateway.EXPECT().StartDetector(ctx, mockDetectorID).Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		detectorID, err := ctrl.CreateAnomalyDetector(ctx, r)
		assert.NoError(t, err)
		assert.NotNil(t, detectorID)
		assert.EqualValues(t, mockDetectorID, *detectorID)
	})
	t.Run("entity failed because of failed to start", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().CreateDetector(ctx, getCreateDetector()).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockADGateway.EXPECT().StartDetector(ctx, mockDetectorID).Return(errors.New("error"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateAnomalyDetector(ctx, r)
		assert.EqualError(t, err, fmt.Sprintf("detector is created with id: %s, but failed to start due to error", mockDetectorID))
	})
}

func TestController_DeleteDetector(t *testing.T) {
	t.Run("invalid detector Id", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, "", false, false)
		assert.EqualError(t, err, "detector Id cannot be empty")
	})
	t.Run("delete gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().DeleteDetector(ctx, mockDetectorID).Return(errors.New("gateway failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, mockDetectorID, false, false)
		assert.EqualError(t, err, "gateway failed")
	})
	t.Run("delete gateway succeeded", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().DeleteDetector(ctx, mockDetectorID).Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, mockDetectorID, false, false)
		assert.NoError(t, err)
	})
	t.Run("stop gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StopDetector(ctx, mockDetectorID).Return(nil, errors.New("failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, mockDetectorID, false, true)
		assert.EqualError(t, err, "failed")
	})
	t.Run("stop gateway succeeded", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StopDetector(ctx, mockDetectorID).Return(mapper.StringToStringPtr("Stopped Detector"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, mockDetectorID).Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, mockDetectorID, false, true)
		assert.NoError(t, err)
	})

	t.Run("cancelled from user", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("no\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, mockDetectorID, true, false)
		assert.NoError(t, err)
	})
	t.Run("agreed by user", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		mockADGateway.EXPECT().DeleteDetector(ctx, mockDetectorID).Return(nil)
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetector(ctx, mockDetectorID, true, false)
		assert.NoError(t, err)
	})
}

func TestController_CreateMultiEntityAnomalyDetector(t *testing.T) {
	t.Run("create one detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter(getRawFilter())
		mockADGateway.EXPECT().CreateDetector(ctx, gatewayRequest).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockADGateway.EXPECT().StartDetector(ctx, mockDetectorID).Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(helperConvertToInterface([]string{"localhost"}), nil)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		detectorID, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, false, false)
		assert.NoError(t, err)
		assert.NotNil(t, detectorID)
		assert.EqualValues(t, gatewayRequest.Name, detectorID[0])
	})
	t.Run("create detector failed due to second detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter(getRawFilter())
		mockADGateway.EXPECT().CreateDetector(ctx, gatewayRequest).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockADGateway.EXPECT().CreateDetector(ctx, gatewayRequest).Return(nil, errors.New(string(helperLoadBytes(t, "create_failed_response.json"))))
		mockADGateway.EXPECT().StartDetector(ctx, mockDetectorID).Return(nil)
		mockADGateway.EXPECT().StopDetector(ctx, mockDetectorID).Return(mapper.StringToStringPtr("stopped"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, mockDetectorID).Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(helperConvertToInterface([]string{"localhost", "localhost"}), nil)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, false, false)
		assert.EqualError(t, err, "Cannot create anomaly detector with name [testdata-detector] as it's already used by detector [wR_1XXMBs3q1IVz33Sk-]")
	})
	t.Run("create one detector with no filter", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		r.Filter = nil
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter()
		mockADGateway.EXPECT().CreateDetector(ctx, gatewayRequest).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockADGateway.EXPECT().StartDetector(ctx, mockDetectorID).Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(helperConvertToInterface([]string{"localhost"}), nil)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		detectorID, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, true, false)
		assert.NoError(t, err)
		assert.NotNil(t, detectorID)
		assert.EqualValues(t, gatewayRequest.Name, detectorID[0])
	})
	t.Run("create one detector interactive rejected", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter(getRawFilter())
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(helperConvertToInterface([]string{"localhost"}), nil)
		var stdin bytes.Buffer
		stdin.Write([]byte("no\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		detectorID, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, true, false)
		assert.NoError(t, err)
		assert.Nil(t, detectorID)
	})
	t.Run("create detector failed since no values in partition field", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter(getRawFilter())
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(nil, nil)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, false, false)
		assert.EqualError(t, err, "failed to get values for partition field: ip, check whether any data is available in index [order*]")
	})
	t.Run("create detector failed since platform controller failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter(getRawFilter())
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(nil, errors.New("failed"))
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, false, false)
		assert.EqualError(t, err, "failed")
	})
	t.Run("create detector failed due to second detector, failed to cleanup automatically", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		r := getCreateDetectorRequest()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		gatewayRequest := getCreateDetector()
		gatewayRequest.Name = gatewayRequest.Name + "-" + "localhost"
		gatewayRequest.Filter = getFinalFilter(getRawFilter())
		mockADGateway.EXPECT().CreateDetector(ctx, gatewayRequest).Return(helperLoadBytes(t, "create_response.json"), nil)
		mockADGateway.EXPECT().CreateDetector(ctx, gatewayRequest).Return(nil, errors.New(string(helperLoadBytes(t, "create_failed_response.json"))))
		mockADGateway.EXPECT().StartDetector(ctx, mockDetectorID).Return(nil)
		mockADGateway.EXPECT().StopDetector(ctx, mockDetectorID).Return(mapper.StringToStringPtr("stopped"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, mockDetectorID).Return(errors.New("failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		mockESController.EXPECT().GetDistinctValues(ctx, r.Index[0], *r.PartitionField).Return(helperConvertToInterface([]string{"localhost", "localhost"}), nil)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.CreateMultiEntityAnomalyDetector(ctx, r, false, false)
		assert.EqualError(t, err, "Cannot create anomaly detector with name [testdata-detector] as it's already used by detector [wR_1XXMBs3q1IVz33Sk-]")
	})
}

func getSearchPayload(name string) entity.SearchRequest {
	return entity.SearchRequest{
		Query: entity.SearchQuery{
			Match: entity.Match{
				Name: name,
			},
		},
	}
}

func TestController_StopDetectorByName(t *testing.T) {
	t.Run("stop empty detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctx := context.Background()
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.StopDetectorByName(ctx, "", false)
		assert.Error(t, err)
	})
	t.Run("stop detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(nil, errors.New("gateway failed"))
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.StopDetectorByName(ctx, "detector", false)
		assert.NoError(t, err)
	})
	t.Run("search detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(nil, errors.New("gateway failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.StopDetectorByName(ctx, "detector", false)
		assert.Error(t, err)
	})
	t.Run("stop detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(mapper.StringToStringPtr("Stopped Detector"), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.StopDetectorByName(ctx, "detector", false)
		assert.NoError(t, err)
	})
}

func TestController_StartDetectorByName(t *testing.T) {
	t.Run("start empty detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctx := context.Background()
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.StartDetectorByName(ctx, "", false)
		assert.Error(t, err)
	})
	t.Run("start detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StartDetector(ctx, "detectorID").Return(errors.New("gateway failed"))
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.StartDetectorByName(ctx, "detector", false)
		assert.NoError(t, err)
	})
	t.Run("search detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(nil, errors.New("gateway failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.StartDetectorByName(ctx, "detector", false)
		assert.Error(t, err)
	})
	t.Run("start detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StartDetector(ctx, "detectorID").Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.StartDetectorByName(ctx, "detector", false)
		assert.NoError(t, err)
	})
}

func TestController_DeleteDetectorByName(t *testing.T) {
	t.Run("invalid detector name", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, "", false, false)
		assert.EqualError(t, err, "name cannot be empty")
	})
	t.Run("delete gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload(mockDetectorName)).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, "detectorID").Return(errors.New("gateway failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, mockDetectorName, false, false)
		assert.NoError(t, err)
	})
	t.Run("delete gateway succeeded", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload(mockDetectorName)).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, "detectorID").Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, mockDetectorName, false, false)
		assert.NoError(t, err)
	})
	t.Run("stop gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload(mockDetectorName)).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(nil, errors.New("failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, mockDetectorName, true, false)
		assert.NoError(t, err)
	})
	t.Run("stop gateway succeeded", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload(mockDetectorName)).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(mapper.StringToStringPtr("Stopped Detector"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, "detectorID").Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, mockDetectorName, true, false)
		assert.NoError(t, err)
	})

	t.Run("cancelled from user", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("no\n"))
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload(mockDetectorName)).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, mockDetectorName, true, false)
		assert.NoError(t, err)
	})
	t.Run("agreed by user", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload(mockDetectorName)).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().StopDetector(ctx, "detectorID").Return(mapper.StringToStringPtr("Stopped Detector"), nil)
		mockADGateway.EXPECT().DeleteDetector(ctx, "detectorID").Return(nil)
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.DeleteDetectorByName(ctx, mockDetectorName, true, false)
		assert.NoError(t, err)
	})
}

func TestController_GetDetectorByName(t *testing.T) {
	detectorOutput := &entity.DetectorOutput{
		ID:          "detectorID",
		Name:        "detector",
		Description: "Test detector",
		TimeField:   "timestamp",
		Index:       []string{"order*"},
		Features: []entity.Feature{
			{
				Name:             "total_order",
				Enabled:          true,
				AggregationQuery: []byte(`{"total_order":{"sum":{"field":"value"}}}`),
			},
		},
		Filter:        []byte(`{"bool" : {"filter" : [{"exists" : {"field" : "value","boost" : 1.0}}],"adjust_pure_negative" : true,"boost" : 1.0}}`),
		Interval:      "5m",
		Delay:         "1m",
		LastUpdatedAt: 1589441737319,
		SchemaVersion: 0,
	}
	t.Run("get empty detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctx := context.Background()
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.GetDetectorsByName(ctx, "", false)
		assert.Error(t, err)
	})
	t.Run("search detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(nil, errors.New("gateway failed"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		_, err := ctrl.GetDetectorsByName(ctx, "detector", false)
		assert.EqualError(t, err, "gateway failed")
	})
	t.Run("search detector gateway returned empty", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return([]byte(`{}`), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		actual, err := ctrl.GetDetectorsByName(ctx, "detector", false)
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})
	t.Run("get detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().GetDetector(ctx, "detectorID").Return(nil, errors.New("gateway failed"))
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(&stdin, mockESController, mockADGateway)
		_, err := ctrl.GetDetectorsByName(ctx, "detector", false)
		assert.EqualError(t, err, "gateway failed")
	})
	t.Run("get detector", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().SearchDetector(ctx, getSearchPayload("detector")).Return(
			helperLoadBytes(t, "search_response.json"), nil)
		mockADGateway.EXPECT().GetDetector(ctx, "detectorID").Return(helperLoadBytes(t, "get_response.json"), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		res, err := ctrl.GetDetectorsByName(ctx, "detector", false)
		assert.NoError(t, err)
		assert.EqualValues(t, *res[0], *detectorOutput)
	})
}

func TestController_UpdateDetector(t *testing.T) {
	input := entity.UpdateDetectorUserInput{
		ID:          "m4ccEnIBTXsGi3mvMt9p",
		Name:        "test-detector",
		Description: "Test detector",
		TimeField:   "timestamp",
		Index:       []string{"order*"},
		Features: []entity.Feature{
			{
				Name:             "total_order",
				Enabled:          true,
				AggregationQuery: []byte(`{"total_order":{"sum":{"field":"value"}}}`),
			},
		},
		Filter:        []byte(`{"bool" : {"filter" : [{"exists" : {"field" : "value","boost" : 1.0}}],"adjust_pure_negative" : true,"boost" : 1.0}}`),
		Interval:      "5m",
		Delay:         "1m",
		LastUpdatedAt: 1589441737319,
		SchemaVersion: 0,
	}
	request := entity.UpdateDetector{
		Name:        "test-detector",
		Description: "Test detector",
		TimeField:   "timestamp",
		Index:       []string{"order*"},
		Features: []entity.Feature{
			{
				Name:             "total_order",
				Enabled:          true,
				AggregationQuery: []byte(`{"total_order":{"sum":{"field":"value"}}}`),
			},
		},
		Filter: []byte(`{"bool" : {"filter" : [{"exists" : {"field" : "value","boost" : 1.0}}],"adjust_pure_negative" : true,"boost" : 1.0}}`),
		Interval: entity.Interval{Period: entity.Period{
			Duration: 5,
			Unit:     "Minutes",
		}},
		Delay: entity.Interval{Period: entity.Period{
			Duration: 1,
			Unit:     "Minutes",
		}},
	}
	t.Run("update detector without ID", func(t *testing.T) {
		invalidInput := input
		invalidInput.ID = ""
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		ctx := context.Background()
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, invalidInput, true, true)
		assert.EqualError(t, err, "detector Id cannot be empty")
	})
	t.Run("stale detector update failure", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().GetDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(helperLoadBytes(t, "get_response.json"), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		staleDetector := input
		staleDetector.LastUpdatedAt = input.LastUpdatedAt - 100
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, staleDetector, false, true)
		assert.EqualError(t, err, "new version for detector is available. Please fetch latest version and then merge your changes")
	})
	t.Run("don't update if user says no", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().GetDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(helperLoadBytes(t, "get_response.json"), nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("no\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, false, true)
		assert.NoError(t, err)
	})
	t.Run("stop detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().StopDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(nil, errors.New("failed to stop detector"))
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, true, true)
		assert.EqualError(t, err, "failed to stop detector")
	})
	t.Run("get detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().GetDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(
			nil, errors.New("failed to get detector"))
		mockESController := mockController.NewMockController(mockCtrl)
		ctrl := New(os.Stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, false, false)
		assert.EqualError(t, err, "failed to get detector")
	})
	t.Run("update detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().GetDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(helperLoadBytes(t, "get_response.json"), nil)
		mockADGateway.EXPECT().UpdateDetector(ctx, "m4ccEnIBTXsGi3mvMt9p", &request).Return(errors.New("failed to update"))
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, false, true)
		assert.EqualError(t, err, "failed to update")
	})
	t.Run("start detector gateway failed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().GetDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(helperLoadBytes(t, "get_response.json"), nil)
		mockADGateway.EXPECT().UpdateDetector(ctx, "m4ccEnIBTXsGi3mvMt9p", &request).Return(nil)
		mockADGateway.EXPECT().StartDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(errors.New("failed to start"))
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, false, true)
		assert.EqualError(t, err, "failed to start")
	})
	t.Run("force detector update success", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockESController := mockController.NewMockController(mockCtrl)
		staleDetector := input
		staleDetector.LastUpdatedAt = input.LastUpdatedAt - 100
		mockADGateway.EXPECT().StopDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(nil, nil)
		mockADGateway.EXPECT().UpdateDetector(ctx, "m4ccEnIBTXsGi3mvMt9p", &request).Return(nil)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, true, false)
		assert.NoError(t, err)
	})
	t.Run("update detector and start", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		ctx := context.Background()
		mockADGateway := gateway.NewMockGateway(mockCtrl)
		mockADGateway.EXPECT().GetDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(helperLoadBytes(t, "get_response.json"), nil)
		mockADGateway.EXPECT().UpdateDetector(ctx, "m4ccEnIBTXsGi3mvMt9p", &request).Return(nil)
		mockADGateway.EXPECT().StartDetector(ctx, "m4ccEnIBTXsGi3mvMt9p").Return(nil)
		mockESController := mockController.NewMockController(mockCtrl)
		var stdin bytes.Buffer
		stdin.Write([]byte("yes\n"))
		ctrl := New(&stdin, mockESController, mockADGateway)
		err := ctrl.UpdateDetector(ctx, input, false, true)
		assert.NoError(t, err)
	})
}
