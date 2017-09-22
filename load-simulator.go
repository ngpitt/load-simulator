package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func client(wg *sync.WaitGroup, stop <-chan struct{}, url string, responses *int32) {
	defer wg.Done()
	for {
		select {
		case <-stop:
			return
		default:
			r, err := http.Get(url)
			if err != nil {
				panic(err.Error())
			}
			r.Body.Close()
			atomic.AddInt32(responses, 1)
		}
	}
}

func stats(wg *sync.WaitGroup, stop <-chan struct{}, responses *int32) {
	defer wg.Done()
	for {
		select {
		case <-stop:
			return
		default:
			fmt.Print("\r", *responses, " responses per second...")
			atomic.SwapInt32(responses, 0)
			time.Sleep(time.Second)
		}
	}
}

func main() {
	url := os.Args[1]
	concurrency, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic("Invalid concurrency.")
	}
	runDuration, err := time.ParseDuration(os.Args[3])
	if err != nil {
		panic("Invalid run duration.")
	}
	sleepDuration, err := time.ParseDuration(os.Args[4])
	if err != nil {
		panic("Invalid sleep duration.")
	}
	for {
		fmt.Print("Beginning ", runDuration.String(), " run...\n")
		stop := make(chan struct{})
		wg := &sync.WaitGroup{}
		wg.Add(concurrency + 1)
		responses := int32(0)
		for i := 0; i < concurrency; i++ {
			go client(wg, stop, url, &responses)
		}
		go stats(wg, stop, &responses)
		time.Sleep(runDuration)
		close(stop)
		wg.Wait()
		fmt.Print("\nSleeping for ", sleepDuration.String(), "...\n")
		time.Sleep(sleepDuration)
	}
}
