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

package main

import (
	"fmt"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"strconv"
	"strings"

	"github.com/apiclarity/apiclarity/plugins/api/client/client"
	"github.com/apiclarity/apiclarity/plugins/api/client/client/operations"
	"github.com/apiclarity/apiclarity/plugins/api/client/models"
)

// TODO:
// check that wasm still working!
// pass the host to Config - not sure I can do this
// Send http call async
// Create a Makefile for kong
// Run kong Makefile from APIClarity Makefile
// Find out how to get destination address (and identify the service as internal) - no need
// Change image of init container to be the pushed image from ghcr registry - changed
// Create a README
// add debug logs
// Fix client.GetPort()

type Config struct {
	apiClient *client.APIClarityPluginsTelemetriesAPI
}

// nolint: deadcode
func New() interface{} {
	cfg := client.DefaultTransportConfig()
	transport := httptransport.New("apiclarity.apiclarity:9000", "/api", cfg.Schemes)
	apiClient := client.New(transport, strfmt.Default)
	return &Config{
		apiClient: apiClient,
	}
}

func (conf Config) Response(kong *pdk.PDK) {
	telemetry, err := createTelemetry(kong)
	if err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to create telemetry. %v", err))
		return
	}

	params := operations.NewPostTelemetryParams().WithBody(telemetry)

	go func() {
		_, err = conf.apiClient.Operations.PostTelemetry(params)
		if err != nil {
			_ = kong.Log.Err(fmt.Sprintf("Failed to post telemetry : %v", err))
		}
	}()
}

func createTelemetry(kong *pdk.PDK) (*models.Telemetry, error) {
	routedService, err := kong.Router.GetService()
	if err != nil {
		return nil, fmt.Errorf("failed to get routed serivce. %v", err)
	}
	clientIP, err := kong.Client.GetIp()
	if err != nil {
		return nil, fmt.Errorf("failed to get client ip. %v", err)
	}
	//clientPort, err := kong.Client.GetPort()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get client port. %v", err)
	//}
	destPort := routedService.Port
	host := routedService.Host

	path, err := kong.Request.GetPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get request path. %v", err)
	}
	reqBody, err := kong.Request.GetRawBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get request body. %v", err)
	}
	resBody, err := kong.ServiceResponse.GetRawBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get response body. %v", err)
	}
	//_ = kong.Log.Err(fmt.Sprintf("path: %v", path))
	//_ = kong.Log.Err(fmt.Sprintf("response body: %v", resBody))
	method, err := kong.Request.GetMethod()
	if err != nil {
		return nil, fmt.Errorf("failed to get request method. %v", err)
	}

	statusCode, err := kong.ServiceResponse.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get response status code. %v", err)
	}
	scheme, err := kong.Request.GetScheme()
	if err != nil {
		return nil, fmt.Errorf("failed to get reuqest scheme. %v", err)
	}
	version, err := kong.Request.GetHttpVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get request http version. %v", err)
	}
	reqHeaders, err := kong.Request.GetHeaders(-1) // default limit of 100 headers
	if err != nil {
		return nil, fmt.Errorf("failed to get request headers. %v", err)
	}
	// Unlike kong.Response.GetHeaders(), this function will only return headers
	// that were present in the response from the Service (ignoring headers added
	// by Kong itself)
	resHeaders, err := kong.ServiceResponse.GetHeaders(-1) // default limit of 100 headers
	if err != nil {
		return nil, fmt.Errorf("failed to get response headers. %v", err)
	}

	telemetry := models.Telemetry{
		DestinationAddress:   ":" + strconv.Itoa(destPort), // TODO not sure we have destination ip
		DestinationNamespace: "",
		Request: &models.Request{
			Common: &models.Common{
				TruncatedBody: false,
				Body:          strfmt.Base64(reqBody),
				Headers:       createHeaders(reqHeaders),
				Version:       fmt.Sprintf("%f", version),
			},
			Host:   parseHost(host),
			Method: method,
			Path:   path,
		},
		RequestID: "", // TODO from where
		Response: &models.Response{
			Common: &models.Common{
				TruncatedBody: false,
				Body:          strfmt.Base64(resBody),
				Headers:       createHeaders(resHeaders),
				Version:       fmt.Sprintf("%f", version),
			},
			StatusCode: strconv.Itoa(statusCode),
		},
		Scheme:        scheme,
		SourceAddress: clientIP + ":80",
	}

	return &telemetry, nil
}

func parseHost(kongHost string) string {
	sp := strings.Split(kongHost, ".")

	if len(sp) < 2 {
		return kongHost
	}

	return sp[0] + "." + sp[1]
}

func createHeaders(headers map[string][]string) []*models.Header {
	ret := []*models.Header{}

	// TODO support multiple values for a header
	for header, values := range headers {
		ret = append(ret, &models.Header{
			Key:   header,
			Value: values[0],
		})
	}
	return ret
}

var Version = "0.2"
var Priority = 1

func main () {
	_ = server.StartServer(New, Version, Priority)
}

