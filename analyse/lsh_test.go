package analyse

import (
	"bufio"
	"log"
	"os"
	"testing"
)

func TestCorpus(t *testing.T) {
	corpus := NewCorpus()
	file, err := os.Open("test.log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		corpus.Insert(scanner.Text())
	}
	corpus.PrintBuckets()
}
