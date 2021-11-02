package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/common/log"
	"strconv"

	"github.com/Kong/go-pdk"

	"github.com/apiclarity/apiclarity/plugins/api/client/client/operations"
	"github.com/apiclarity/apiclarity/plugins/api/client/client"
	"github.com/apiclarity/apiclarity/plugins/api/client/models"
	//"github.com/apiclarity/apiclarity/plugins/api/server/restapi/operations"
)

type Config struct {
	apiClient *client.APIClarityPluginsTelemetriesAPI
}

func New() interface{} {
	cfg := client.DefaultTransportConfig()
	cfg.WithHost("apiclarity.apiclarity:8080") // TODO configure it
	transport := httptransport.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	apiClient := client.New(transport, strfmt.Default)
	return &Config{
		apiClient: apiClient,
	}
}

func (conf Config) Response(kong *pdk.PDK) {
	kong.Log.Err("Response")

	telemetry, err := createTelemetry(kong)
	if err != nil {
		log.Errorf("Failed to create telemetry. %v", err)
		return
	}

	params := operations.NewPostTelemetryParams().WithBody(telemetry)

	// TODO handle response of the async call?
	go conf.apiClient.Operations.PostTelemetry(params)

	return
}

func createTelemetry(kong *pdk.PDK) (*models.Telemetry, error) {
	clientIP, err := kong.Client.GetIp()
	if err != nil {
		return nil, fmt.Errorf("failed to get client ip. %v", err)
	}
	clientPort, err := kong.Client.GetPort()
	if err != nil {
		return nil, fmt.Errorf("failed to get client port. %v", err)
	}
	destPort, err := kong.Request.GetPort()
	if err != nil {
		return nil, fmt.Errorf("failed to get destination port. %v", err)
	}
	host, err := kong.Request.GetHost()
	if err != nil {
		return nil, fmt.Errorf("failed to get request host. %v", err)
	}
	reqBody, err := kong.Request.GetRawBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get request body. %v", err)
	}
	resBody, err := kong.ServiceResponse.GetRawBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get response body. %v", err)
	}
	method, err := kong.Request.GetMethod()
	if err != nil {
		return nil, fmt.Errorf("failed to get request method. %v", err)
	}
	path, err := kong.Request.GetPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get request path. %v", err)
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
		Request:              &models.Request{
			Common: &models.Common{
				TruncatedBody: false,
				Body:          reqBody,
				Headers:       createHeaders(reqHeaders),
				Version:       fmt.Sprintf("%f", version),
			},
			Host:   host,
			Method: method,
			Path:   path,
		},
		RequestID:            "", // TODO from where
		Response:             &models.Response{
			Common:     &models.Common{
				TruncatedBody: false,
				Body:          resBody,
				Headers:       createHeaders(resHeaders),
				Version:       fmt.Sprintf("%f", version),
			},
			StatusCode: strconv.Itoa(statusCode),
		},
		Scheme:               scheme,
		SourceAddress:        clientIP+":"+strconv.Itoa(clientPort),
	}

	return &telemetry, nil
}

func createHeaders(headers map[string][]string) []*models.Header {
	ret := []*models.Header{}

	// TODO support multiple values for a header
	for header, values := range headers {
		ret =  append(ret, &models.Header{
			Key:   header,
			Value: values[0],
		})
	}
	return ret
}

//const Version = "1.0.0"
//const Priority = 1

//func main() {
//	if err := server.StartServer(New, Version, Priority); err != nil {
//		fmt.Printf("error staring server: %v\n", err)
//	}
//}


