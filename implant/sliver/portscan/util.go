package portscan

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

func Usage() string {
	msg := `
Usage: portscan -h hostspec -p portspec -t threads [default 8]

Examples:
  portscan -h 10.0.0.0/24 -p 21-23,389,3389
  portscan -h vcenter01.dbg.local -p 443
  portscan -h hosts.txt -p 135,139,445 -t 16
`
	return msg
}

func parseCmdLine() {
	flag.StringVar(&config.hostSpec, "h", "", "Host specification")
	flag.StringVar(&config.portSpec, "p", "", "Port specification")
	flag.IntVar(&config.numThreads, "t", config.numThreads, "Number of worker threads")
	flag.Parse()

	if config.hostSpec == "" || config.portSpec == "" {
		Usage()
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(config.hostSpec)
	if err == nil {
		return true
	} else {
		return false
	}
}

func parseHostSpec() []string {
	var ret []string

	if fileExists(config.hostSpec) {
		file, err := os.Open(config.hostSpec)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ret = append(ret, strings.Trim(scanner.Text(), " "))
		}
	} else if strings.Contains(config.hostSpec, "/") {
		for _, host := range explodeCidr(config.hostSpec) {
			ret = append(ret, host.String())
		}
	} else {
		ret = append(ret, config.hostSpec)
	}

	return ret
}

func atoi(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func parsePortSpec() []int {
	var ports []int

	for _, commas := range strings.Split(config.portSpec, ",") {
		if strings.Contains(commas, "-") {
			dashes := strings.Split(commas, "-")
			start, _ := strconv.Atoi(dashes[0])
			end, _ := strconv.Atoi(dashes[1])
			for i := start; i < end+1; i++ {
				ports = append(ports, i)
			}
		} else {
			ports = append(ports, atoi(commas))
		}
	}

	return ports
}
