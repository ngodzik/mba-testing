package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	rmqdriver "github.com/ngodzik/mba-rabbitmq-driver"
)

func main() {
	var ocNb int
	var producersNb int
	var seed int64

	flag.IntVar(&ocNb, "n", 10, "Sets the number of artificial neural network output computations to be carried out")
	flag.IntVar(&producersNb, "p", 1, "Sets the number of workers producing the inputs")
	flag.Int64Var(&seed, "s", 1, "Sets the seed, -1 or less to set it randomly")

	flag.Parse()

	fmt.Printf("computations number: %d, producers number: %d\n", ocNb, producersNb)

	if seed < 0 {
		seed = time.Now().UTC().UnixNano()
	}

	rand.Seed(seed)

	// Create the message broker

	url := "amqp://guest:guest@localhost:5672/"
	mb, err := rmqdriver.New(url)

	if err != nil {
		panic("Cannot get a new message broker")
	}

	var wgInit sync.WaitGroup
	var wgWorkersP sync.WaitGroup

	type neurons struct {
		Values [2]float64 `json:"values"`
	}

	// Create producers
	for i := 0; i < producersNb; i++ {
		wgInit.Add(1)
		wgWorkersP.Add(1)
		go func(id int) {

			defer wgWorkersP.Done()

			pb, err := mb.NewPublisher("ann")

			if err != nil {
				panic("Cannot get a new publisher")
			}

			defer pb.Close()

			goID := id

			// Initialization done
			wgInit.Done()

		Working:
			for li := 0; li < ocNb; li++ {
				ns := neurons{}

				ns.Values[0] = rand.Float64()
				ns.Values[1] = rand.Float64()

				err := pb.Send(ns)
				if err != nil {
					break Working
				}

			}
			fmt.Printf("Producer: %d, end of goroutine\n", goID)
		}(i)
	}

	// Wait for all goroutines to be initialized
	wgInit.Wait()

	wgWorkersP.Wait()

}
