// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package traces

import (
	"github.com/apiclarity/apiclarity/plugins/api/server/models"
	"github.com/apiclarity/apiclarity/plugins/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/plugins/api/server/restapi"
	"github.com/apiclarity/speculator/pkg/spec"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"

	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/backend/pkg/common"
)

type HandleTraceFunc func(trace *spec.SCNTelemetry) error

type HTTPTracesServer struct {
	traceHandleFunc HandleTraceFunc
	server          *restapi.Server
}

func CreateHTTPTracesServer(port int, traceHandleFunc HandleTraceFunc) (*HTTPTracesServer, error) {
	s := &HTTPTracesServer{}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewAPIClarityPluginsTelemetriesAPIAPI(swaggerSpec)

	api.PostTelemetryHandler = operations.PostTelemetryHandlerFunc(func(params operations.PostTelemetryParams) middleware.Responder {
		return s.PostTelemetry(params)
	})

	server := restapi.NewServer(api)
	defer server.Shutdown()

	server.ConfigureFlags()
	server.ConfigureAPI()

	server.Port = port
	s.server = server
	s.traceHandleFunc = traceHandleFunc

	//if err := server.Serve(); err != nil {
	//	log.Fatalln(err)
	//}

	return s, nil
}

func (s *HTTPTracesServer) Start(errChan chan struct{}) {
	log.Infof("Starting REST server")
	go func() {
		if err := s.server.Serve(); err != nil {
			log.Errorf("Failed to serve REST server: %v", err)
			errChan <- common.Empty
		}
	}()
	log.Infof("REST server is running")
}

func (s *HTTPTracesServer) Stop() {
	log.Infof("Stopping REST server")
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			log.Errorf("Failed to shutdown REST server: %v", err)
		}
	}
}


func (s *HTTPTracesServer) PostTelemetry(params operations.PostTelemetryParams) middleware.Responder {

	telemetry := getTelemetry(params.Body)

	if err := s.traceHandleFunc(telemetry); err != nil {
		// TODO handle error
		log.Errorf("Error from trace handling func: %v", err)
	}

	return operations.NewPostTelemetryOK().WithPayload(&models.SuccessResponse{
		Message: "Success",
	})
}

func getTelemetry(telemetry *models.Telemetry) *spec.SCNTelemetry {
	return &spec.SCNTelemetry{
		RequestID:            telemetry.RequestID,
		Scheme:               telemetry.Scheme,
		DestinationAddress:   telemetry.DestinationAddress,
		DestinationNamespace: telemetry.DestinationNamespace,
		SourceAddress:        telemetry.SourceAddress,
		SCNTRequest:          spec.SCNTRequest{
			Method:     telemetry.Request.Method,
			Path:       telemetry.Request.Path,
			Host:       telemetry.Request.Host,
			SCNTCommon: convertCommon(telemetry.Request.Common),
		},
		SCNTResponse:         spec.SCNTResponse{
			StatusCode: telemetry.Response.StatusCode,
			SCNTCommon: convertCommon(telemetry.Response.Common),
		},
	}
}

func convertCommon(common *models.Common) spec.SCNTCommon {
	return spec.SCNTCommon{
		Version:       common.Version,
		Headers:       convertHeaders(common.Headers),
		Body:          []byte(common.Body),
		TruncatedBody: common.TruncatedBody,
	}
}

func convertHeaders(headers []*models.Header) [][2]string {
	var ret [][2]string

	for _, header := range headers {
		ret = append(ret, [2]string{
			header.Key, header.Value,
		})
	}
	return ret
}

//
//func readHTTPTraceBodyData(req *http.Request) (*spec.SCNTelemetry, error) {
//	decoder := json.NewDecoder(req.Body)
//	var bodyData *spec.SCNTelemetry
//	err := decoder.Decode(&bodyData)
//	if err != nil {
//		return nil, fmt.Errorf("failed to decode trace: %v", err)
//	}
//
//	return bodyData, nil
//}
//
//func (s *HTTPTracesServer) httpTracesHandler(w http.ResponseWriter, r *http.Request) {
//	trace, err := readHTTPTraceBodyData(r)
//	if err != nil || trace == nil {
//		log.Errorf("Invalid trace. err=%v, trace=%+s", err, r.Body)
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	traceB, _ := json.Marshal(trace)
//	log.Infof("Trace was received: %s", traceB)
//	err = s.traceHandleFunc(trace)
//	if err != nil {
//		log.Errorf("Failed to handle trace. err=%v", err)
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	log.Infof("Trace was handled successfully")
//	w.WriteHeader(http.StatusAccepted)
//}
