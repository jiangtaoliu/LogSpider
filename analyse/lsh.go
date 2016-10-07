package analyse

import (
	"fmt"
	"github.com/xrash/smetrics"
)

const THRESHOLD = 0.8

type Bucket struct {
	Common string
	Sample string
}

type Corpus struct {
	Buckets []*Bucket
}

func NewCorpus() *Corpus {
	return &Corpus{Buckets: []*Bucket{}}
}

func CompareStrings(a, b string) float64 {
	val := smetrics.JaroWinkler(a, b, 0.7, 4)
	return val
}

func (this *Corpus) Insert(entry string) (*Bucket, bool) {
	for _, bucket := range this.Buckets {
		if CompareStrings(bucket.Sample, entry) > THRESHOLD {
			bucket.Sample = entry
			return bucket, false
		}
	}
	thisBucket := &Bucket{Sample: entry}
	this.Buckets = append(this.Buckets, thisBucket)
	return thisBucket, true
}

func (this *Corpus) PrintBuckets() {
	fmt.Printf("Total of %d buckets\n", len(this.Buckets))
	for i, bucket := range this.Buckets {
		fmt.Printf("Bucket %d: %s\n", i, bucket.Sample)
	}
}
