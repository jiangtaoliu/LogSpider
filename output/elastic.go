package output

import (
	"github.com/bahusvel/NetworkScannerThingy/logs"
	"gopkg.in/olivere/elastic.v3"
)

type ElasticOutput struct {
	client    *elastic.Client
	IndexName string
	ServerURL string
}

func (this *ElasticOutput) Init() error {
	tmpClient, err := elastic.NewClient(elastic.SetURL(this.ServerURL))
	if err != nil {
		return err
	}
	this.client = tmpClient
	exists, err := this.client.IndexExists(this.IndexName).Do()
	if err != nil {
		return err
	}
	if !exists {
		_, err = this.client.CreateIndex(this.IndexName).Do()
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *ElasticOutput) SendLogEntry(entry logs.LogEntry) error {
	_, err := this.client.Index().Index(this.IndexName).Type("logentry").BodyJson(entry).Refresh(true).Do()
	if err != nil {
		return err
	}
	return nil
}
