package logs_analysis_module

import (
	"gopkg.in/olivere/elastic.v5"
)

type elasticClient struct {
	// es
	elasticClient *elastic.Client
	// es index
	esIndex       string
	// es type
	esType        string
}

func New(elasticSearchUrl, esIndex, esType string) (*elasticClient, error) {
	var esClient *elasticClient
	es, err := elastic.NewClient(elastic.SetURL(elasticSearchUrl),
		elastic.SetSniff(false))
	if err == nil {
		if esIndex == "" {
			esIndex = "eospark_web-"
		}

		if esType == "" {
			esType = "api"
		}
		esClient = &elasticClient{
			elasticClient: es,
			esIndex:       esIndex,
			esType:        esType,
		}
	}
	return esClient, err
}
