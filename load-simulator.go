package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/autoscaling/v1"
	"k8s.io/client-go/rest"
)

func getHpa(name string, client *kubernetes.Clientset) *v1.HorizontalPodAutoscaler {
	hpa, err := client.AutoscalingV1().HorizontalPodAutoscalers(metav1.NamespaceDefault).Get(name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	return hpa
}

func main() {
	url := flag.String("url", "", "target url")
	numClients := flag.Int("clients", 8, "number of clients")
	hpaName := flag.String("hpa", "", "target app hpa")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	flag.Parse()
	for {
		hpa := getHpa(*hpaName, client)
		log.Print("Targeting ", *hpa.Spec.TargetCPUUtilizationPercentage, "% CPU utilization with ", *numClients, " clients...\n")
		stopCh := make(chan struct{})
		wg := sync.WaitGroup{}
		wg.Add(*numClients)
		for i := 0; i < *numClients; i++ {
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stopCh:
						return
					default:
					}
					r, err := http.Get(*url)
					if err != nil {
						panic(err.Error())
					}
					r.Body.Close()
				}
			}()
		}
		rampingUp := true
		for {
			time.Sleep(5 * time.Second)
			hpa := getHpa(*hpaName, client)
			if hpa.Status.CurrentCPUUtilizationPercentage == nil {
				log.Print("Waiting for CPU utilization data...\n")
			} else {
				log.Print(*hpa.Status.CurrentCPUUtilizationPercentage, "% current CPU utilization...\n")
				if rampingUp {
					if *hpa.Status.CurrentCPUUtilizationPercentage > *hpa.Spec.TargetCPUUtilizationPercentage+5 {
						rampingUp = false
					}
				} else if *hpa.Status.CurrentCPUUtilizationPercentage <= *hpa.Spec.TargetCPUUtilizationPercentage+5 {
					log.Print(*hpa.Spec.TargetCPUUtilizationPercentage, "% CPU utilization reached, waiting for scale down...\n")
					break
				}
			}
		}
		close(stopCh)
		wg.Wait()
		for {
			time.Sleep(5 * time.Second)
			hpa := getHpa(*hpaName, client)
			if hpa.Status.CurrentReplicas == *hpa.Spec.MinReplicas {
				break
			}
		}
	}
}
