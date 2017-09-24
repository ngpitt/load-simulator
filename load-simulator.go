package main

import (
	"flag"
	"log"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "", "target url")
	clients := flag.Int("clients", 8, "number of clients")
	runDuration := flag.Duration("run_duration", 2*time.Minute, "run duration")
	sleepDuration := flag.Duration("sleep_duration", 30*time.Second, "sleep duration")
	flag.Parse()
	for {
		log.Println("Beginning", runDuration.String(), "run...")
		stop := make(chan struct{})
		wg := sync.WaitGroup{}
		wg.Add(*clients + 1)
		responses := int32(0)
		for i := 0; i < *clients; i++ {
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
					}
					r, err := http.Get(*url)
					if err != nil {
						panic(err.Error())
					}
					r.Body.Close()
					atomic.AddInt32(&responses, 1)
				}
			}()
		}
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				log.Println(math.Floor(float64(atomic.LoadInt32(&responses))/5+0.5), "responses per second...")
				atomic.SwapInt32(&responses, 0)
				time.Sleep(5 * time.Second)
			}
		}()
		time.Sleep(*runDuration)
		close(stop)
		wg.Wait()
		log.Println("Sleeping for", sleepDuration.String(), "...")
		time.Sleep(*sleepDuration)
	}
}
