package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

func testURL(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")

	// Get the query parameters
	params := r.URL.Query()

	// Get the URL parameter
	urlParam, ok := params["url"]
	if !ok || len(urlParam[0]) < 1 {
		http.Error(w, "Url Param 'url' is missing", http.StatusBadRequest)
		return
	}

	url := urlParam[0]
	log.Println("URL: ", url)

	// Get the number parameter
	numberParam, ok := params["threads"]
	if !ok || len(numberParam[0]) < 1 {
		http.Error(w, "Url Param 'number' is missing", http.StatusBadRequest)
		return
	}

	concurrentParam, ok := params["concurrent"]
	if !ok || len(concurrentParam[0]) < 1 {
		concurrentParam = numberParam
	}

	// Convert the threads parameter to an integer
	threads, err := strconv.Atoi(numberParam[0])
	if err != nil {
		log.Println("Url Param 'threads' must be an integer")
		http.Error(w, "Url Param 'threads' must be an integer", http.StatusInternalServerError)
		return
	}

	concurrent, err := strconv.Atoi(concurrentParam[0])
	if err != nil {
		log.Println("Url Param 'concurrent' must be an integer")
		http.Error(w, "Url Param 'concurrent' must be an integer", http.StatusInternalServerError)
		return
	}

	success, errors := testURLs(url, threads, concurrent)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Success: %d, Errors: %d", success, errors)))
}

// testURLs sends multiple GET requests to the specified URL concurrently.
// It takes in the URL to send the requests to, the number of requests to send,
// and the maximum number of concurrent requests to make.
func testURLs(url string, threads, concurrent int) (int, int) {
	channel := make(chan int, concurrent)
	codes := make(chan int, threads)
	requestsSuccess := 0
	requestsError := 0

  var wg sync.WaitGroup
  wg.Add(threads)
  
	for i := 0; i < threads; i++ {
		channel <- i
		go func(){
      makeGetRequest(url, channel, codes)
      wg.Done()
      }()
	}
	close(channel)

  go func(){
    wg.Wait()
    close(codes)
  }()

	for code := range codes {
		log.Println("Received status code: ", code)
		// <-channel
		if code != http.StatusOK {
			requestsError++
		} else {
			requestsSuccess++
		}
	}

	log.Printf("Sent %d requests to %s\n", requestsSuccess, url)
	return requestsSuccess, requestsError
}

func makeGetRequest(url string, ch, codesChan chan int) {
	for range ch {
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error making GET request: ", err)
			codesChan <- http.StatusInternalServerError
		}
		defer resp.Body.Close()

		codesChan <- resp.StatusCode
	}
}

func initHttpServer(port int) {
	http.HandleFunc("/test-url", testURL)
	log.Printf("Server started on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func main() {
	initHttpServer(8080)
}
