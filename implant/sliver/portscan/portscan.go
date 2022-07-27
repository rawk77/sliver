package portscan

import (
	"sync"
)

type Config struct {
	hostSpec   string
	portSpec   string
	numThreads int
}

var config Config

func initConfig() {
	config.numThreads = 8
}

func Scan(HostSpec string, PortSpec string) string {

	config.hostSpec = HostSpec
	config.portSpec = PortSpec

	initConfig()

	var probes []*Probe
	var output string

	for _, host := range parseHostSpec() {
		for _, port := range parsePortSpec() {
			probes = append(probes, NewProbe(host, port))
		}
	}

	input := make(chan *Probe, config.numThreads)
	results := make(chan *Probe)

	var wgProducers sync.WaitGroup
	var wgConsumers sync.WaitGroup
	wgProducers.Add(config.numThreads)
	wgConsumers.Add(1)

	for i := 0; i < config.numThreads; i++ {
		go func() {
			defer wgProducers.Done()

			for {
				probe := <-input
				if probe == nil {
					break
				}

				probe.Probe()
				results <- probe
			}
		}()
	}

	go func() {
		defer wgConsumers.Done()
		for result := range results {
			if result.open == true {
				output += result.Report()
			}
		}
	}()

	for _, probe := range probes {
		input <- probe
	}

	close(input)
	wgProducers.Wait()
	close(results)
	wgConsumers.Wait()

	return output
}
