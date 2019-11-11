package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// SingleHash - crc32(data) + "~" + crc32(md5(data)
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for dataRaw := range in {
		dataNum := dataRaw.(int)
		data := strconv.Itoa(dataNum)

		wg.Add(1)
		md5 := DataSignerMd5(data)
		crc32Ch := make(chan string, 1)
		go func(data string) {
			defer wg.Done()

			go func(ch chan string) {
				ch <- DataSignerCrc32(data)
			}(crc32Ch)
			crc32fromMd5 := DataSignerCrc32(md5)

			out <- (<-crc32Ch) + "~" + crc32fromMd5
		}(data)
	}

	wg.Wait()
}

// MultiHash - crc32(th + data), th = 0..5, concat
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for dataRaw := range in {
		data := dataRaw.(string)

		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			channels := make([]chan string, 6)
			for th := 0; th < 6; th++ {
				channels[th] = make(chan string, 1)

				go func(ch chan string, input string) {
					ch <- DataSignerCrc32(input)
				}(channels[th], strconv.Itoa(th)+data)
			}

			var result string
			for _, ch := range channels {
				result += <-ch
			}

			out <- result
		}(data)
	}

	wg.Wait()
}

// CombineResults - join(sort(all_data), "_")
func CombineResults(in, out chan interface{}) {
	results := []string{}
	for dataRaw := range in {
		data := dataRaw.(string)
		results = append(results, data)
	}

	sort.Strings(results)
	out <- strings.Join(results, "_")
}

// ExecutePipeline - unix pipeline analog
func ExecutePipeline(jobs ...job) {
	runWorker := func(
		worker job,
		in, out chan interface{},
		waiter *sync.WaitGroup,
	) {
		worker(in, out)
		close(out)
		waiter.Done()
	}

	var in, out chan interface{}
	wg := &sync.WaitGroup{}

	for i := 0; i < len(jobs); i++ {
		wg.Add(1)
		out = make(chan interface{})
		go runWorker(jobs[i], in, out, wg)

		in = out
	}
	wg.Wait()
}
