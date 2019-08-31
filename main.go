package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	config := parseFlags()

	var wg sync.WaitGroup
	var resultChannels []chan testResult
	for i := 0; i < config.iterations; i++ {
		wg.Add(1)
		// result channel with buffer size of 1
		// we need a buffer of 1 so that we don't block the goroutines completion
		var resultChannel = make(chan testResult, 1)
		resultChannels = append(resultChannels, resultChannel)
		go func(wg *sync.WaitGroup, result chan testResult, url string) {
			defer wg.Done()
			startTime := time.Now()
			r, e := http.Get(url)
			endTime := time.Since(startTime)
			if e != nil {
				log.Printf("Error: %v\n", e)
			}

			result <- testResult{
				Duration: endTime.Seconds(),
				Header:   r.Header,
			}
			// channels are reference types and don't need to be passed as pointers
		}(&wg, resultChannel, config.url)
	}

	// we wait this goroutine and that starts all the others until wg is complete
	wg.Wait()

	var totalExecutionTimeNanos float64
	var header http.Header
	for _, channel := range resultChannels {
		tResult := <-channel
		totalExecutionTimeNanos += tResult.Duration
		header = tResult.Header
	}
	fmt.Print(header)
	fmt.Printf("Result iterations=%d avgExecutionTime=%f(seconds)\n", config.iterations, totalExecutionTimeNanos/float64(config.iterations))
}

type testResult struct {
	Duration float64
	Header   http.Header
}

type mainConfig struct {
	iterations int
	url        string
}

func parseFlags() mainConfig {
	iterations := flag.Int("iterations", 10, "[default 10] number of times to hit the URL.")
	url := flag.String("url", "", "[default nil] URL endpoint to hit.")
	flag.Parse()
	return mainConfig{
		iterations: *iterations,
		url:        *url,
	}
}
