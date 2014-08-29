package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

//   1) Read URL from file      (urlReader)
//   2) Fetch resource          (resourceFetcher)
//   3) Consume resource        (resourceConsumer)

type HttpPair struct {
	Req  *http.Request
	Resp *http.Response
	Time time.Duration // time from request to receiving the response header
}

// read URLs from file, pass to channel
func urlReader(file string) (<-chan string, error) {

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	ch := make(chan string)
	go func() {
		defer f.Close()
		for scanner.Scan() {
			ch <- scanner.Text()
		}

		close(ch)
	}()

	return ch, nil
}

// read urls from channel, fetch HTTP response and pass to channel
func resourceFetcher(cli *http.Client, urls <-chan string) <-chan *HttpPair {
	ch := make(chan *HttpPair)
	go func() {
		for url := range urls {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				continue
			}

			start := time.Now()
			resp, err := cli.Do(req)
			t := time.Since(start)
			if err == nil {
				ch <- &HttpPair{Req: req, Resp: resp, Time: t}
			}
		}

		close(ch)
	}()

	return ch
}

// read HTTP request/response-pair from channel and do something with it (print it)
func resourceConsumer(msgs <-chan *HttpPair, consumeBody bool) {
	for msg := range msgs {
		if consumeBody {
			ioutil.ReadAll(msg.Resp.Body)
		}

		fmt.Printf("%v %v %v\n", msg.Req.URL, msg.Resp.Status, msg.Time)
		msg.Resp.Body.Close()
	}
}

func mergeChans(ms []<-chan *HttpPair) <-chan *HttpPair {
	var wg sync.WaitGroup
	msgChan := make(chan *HttpPair)
	output := func(c <-chan *HttpPair) {
		for v := range c {
			msgChan <- v
		}

		wg.Done()
	}

	wg.Add(len(ms))
	for _, m := range ms {
		go output(m)
	}

	go func() {
		wg.Wait()
		close(msgChan)
	}()

	return msgChan
}

func main() {
	var urlFile = flag.String("url-file", "", "file containing a newline separated list of URLs")
	var nfetchers = flag.Int("n-fetchers", 20, "number of concurrent HTTP fetchers")
	var httpTimeout = flag.Duration("http-timeout", 20*time.Second, "HTTP client timeout")
	var consumeBody = flag.Bool("consume-body", false, "consume http response body")
	flag.Parse()

	cli := &http.Client{Timeout: *httpTimeout}
	if len(*urlFile) == 0 {
		flag.PrintDefaults()
		return
	}

	urls, err := urlReader(*urlFile)
	if err != nil {
		fmt.Println(err)
	}

	var ms []<-chan *HttpPair
	for i := 0; i < *nfetchers; i++ {
		m := resourceFetcher(cli, urls)
		ms = append(ms, m)
	}

	msgChan := mergeChans(ms)
	resourceConsumer(msgChan, *consumeBody)
}
