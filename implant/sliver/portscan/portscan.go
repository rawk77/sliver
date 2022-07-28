package portscan

import (
	"fmt"
	"sync"
)

type Config struct {
	hostSpec   string
	portSpec   string
	numThreads int32
}

var config Config

func Scan(hostSpec string, portSpec string, numThreads int32) (string, error) {
	config.hostSpec = hostSpec
	config.portSpec = portSpec
	config.numThreads = numThreads

	var probes []*Probe
	var output string

	for _, host := range parseHostSpec() {
		for _, port := range parsePortSpec() {
			probes = append(probes, NewProbe(host, port))
		}
	}

	if len(probes) == 0 {
		return "", fmt.Errorf("No host/port pairs could be loaded")
	}

	input := make(chan *Probe, config.numThreads)
	results := make(chan *Probe)

	var wgProducers sync.WaitGroup
	var wgConsumers sync.WaitGroup
	wgProducers.Add(int(config.numThreads))
	wgConsumers.Add(1)

	for i := 0; i < int(config.numThreads); i++ {
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
				output += result.Report() + "\n"
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

	if output == "" {
		output = "No open ports were found"
	}

	return output, nil
}
