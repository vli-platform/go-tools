package elastic

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type ElasticCLuster struct {
	CloudId  string
	Host     []string
	Username string
	Password string
}

func New(cfg ElasticCLuster) (*elasticsearch.Client, error) {
	// tham khảo tại https://www.elastic.co/blog/the-go-client-for-elasticsearch-working-with-data
	retryBackoff := backoff.NewExponentialBackOff()
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		CloudID:   cfg.CloudId,
		Addresses: cfg.Host,
		Username:  cfg.Username,
		Password:  cfg.Password,
		// Retry on 429 TooManyRequests statuses
		RetryOnStatus: []int{502, 503, 504, 429},
		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		MaxRetries: 5,

		Transport: &http.Transport{
			MaxIdleConnsPerHost:   20,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("elasticsearch.NewClient is error: %s", err)
	}
	return es, nil
}

type EsQueryBuilder struct {
	Source []string    `json:"_source,omitempty"`
	Query  *EsQuery    `json:"query,omitempty"`
	Size   int         `json:"size,omitempty"`
	Sort   interface{} `json:"sort,omitempty"`
}

type EsQuery struct {
	Bool          interface{}  `json:"bool,omitempty"`
	Must          interface{}  `json:"must,omitempty"`
	Match         interface{}  `json:"match,omitempty"`
	MatchAll      interface{}  `json:"match_all,omitempty"`
	Range         interface{}  `json:"range,omitempty"`
	MustNot       interface{}  `json:"must_not,omitempty"`
	Terms         interface{}  `json:"terms,omitempty"`
	Term          interface{}  `json:"term,omitempty"`
	FunctionScore interface{}  `json:"function_score,omitempty"`
	Query         interface{}  `json:"query,omitempty"`
	RandomScore   *RandomScore `json:"random_score,omitempty"`
	BoostMode     string       `json:"boost_mode,omitempty"`
}

type RandomScore struct {
	Seed int `json:"seed,omitempty"`
}

func Match(query interface{}) *EsQuery {
	return &EsQuery{Match: query}
}
func Range(query interface{}) *EsQuery {
	return &EsQuery{Range: query}
}
func Terms(query interface{}) *EsQuery {
	return &EsQuery{Terms: query}
}

func EsOnSuccess(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
}
func EsOnFail(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
	log.Printf("Elasticsearch / Fail to %s on index: %s, docId: %s, error: %s\n", item.Action, item.Index, item.DocumentID, res.Error.Reason)
}
