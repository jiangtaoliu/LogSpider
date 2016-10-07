package analyse

import "github.com/bahusvel/NetworkScannerThingy/logs"

type Logs map[string]*Corpus

type KnowledgeDB map[string]*Logs

func (this KnowledgeDB) Classify(entry logs.LogEntry) (*Bucket, bool) {
	hostLogs, ok := this[entry.Host]
	if !ok {
		tmpLogs := &Logs{}
		this[entry.Host] = tmpLogs
		hostLogs = tmpLogs
	}
	logCorpus, ok := (*hostLogs)[entry.Log]
	if !ok {
		tmpCorpus := NewCorpus()
		(*hostLogs)[entry.Log] = tmpCorpus
		logCorpus = tmpCorpus
	}
	return logCorpus.Insert(entry.Entry)
}
