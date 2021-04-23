package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func main() {

	// Sample usage

	whichMetric := "CapRealUSD"
	sinceWhen := yesterdaySimpleDate() // "" will start with earliest possible datapoint
	UntilWhen := ""                    // most recent datapoint

	opts := CMAPIListOptions{Metrics: whichMetric, Start: sinceWhen, End: UntilWhen}
	md, err := GetMetricData(context.Background(), *NewCommunityClient(""), &opts)
	if err != nil {
		panic(err)
	}

	res := fmt.Sprintf(
		"latest %s : %s @ %s",
		whichMetric,
		md.Data.Series[0].Values[0],
		md.Data.Series[0].Date,
	)

	println(res)
	// latest CapRealUSD : 357839722701.057414591381506310739636 @ 2021-04-22T00:00:00.000Z
}

const (
	CommunityBaseURLV3 = "https://community-api.coinmetrics.io/v3/"
	MetricPathV3       = "assets/btc/metricdata"
)

// client is a generic HTTP structure serving as a base for communication
type client struct {
	BaseURL string
	apiKey  string
	http    *http.Client
}

// NewCommunityClient is a starting point for getting a handle on CM API
// As of API v3 you dont need an API key, just pass ""
func NewCommunityClient(apiKey string) *client {
	return &client{
		BaseURL: CommunityBaseURLV3,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// CMAPIListOptions lets you wrap API query parameters
// Should be passed over into the GetMetricData function
type CMAPIListOptions struct {
	Metrics string
	Start   string
	End     string
}

func (c *client) SendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetMetricData is the main func used to get data from CM, arguments:
// api can be instantiated with NewCommunityClient(""),
// options can be instantiated as a literal from CMAPIListOptions
func GetMetricData(ctx context.Context, api client, options *CMAPIListOptions) (*MetricData, error) {

	// -------------
	// Bootstrap URL
	// -------------

	baseURL, err := url.Parse(api.BaseURL)
	if err != nil {
		fmt.Println("GetMetricData malformed URL: ", err.Error())
		return nil, err
	}
	baseURL.Path += MetricPathV3

	// --------------
	// URL parameters
	// --------------

	params := url.Values{}

	params.Add("metrics", options.Metrics)

	if options.Start != "" {
		params.Add("start", options.Start)
	}

	if options.End != "" {
		params.Add("end", options.End)
	}

	baseURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// -----------------
	// Request execution
	// -----------------

	req = req.WithContext(ctx)

	res := MetricData{}

	if err := api.SendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// --------------
// Basic mappings
// --------------

// Sample response from CoinMetrics

// {
//     "metricData": {
//         "metrics": [
//             "CapRealUSD"
//         ],
//         "series": [
//             {
//                 "time": "2021-04-22T00:00:00.000Z",
//                 "values": [
//                     "357839722701.057414591381506310739636"
//                 ]
//             }
//         ]
//     }
// }

type MetricData struct {
	Data DataSeries `json:"metricData"`
}

type DataSeries struct {
	Metrics []string    `json:"metrics"`
	Series  []DataPoint `json:"series"`
}

type DataPoint struct {
	Date   string   `json:"time"`
	Values []string `json:"values"`
}

// -----
// Utils
// -----

func yesterdaySimpleDate() string {
	dt := time.Now().Add(-24 * time.Hour)
	return fmt.Sprintf("%d-%02d-%02d", dt.Year(), dt.Month(), dt.Day())
}
