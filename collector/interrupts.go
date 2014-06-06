// +build !nointerrupts

package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	procInterrupts = "/proc/interrupts"
)

var (
	interruptsMetric = prometheus.NewCounter()
)

type interruptsCollector struct {
	registry prometheus.Registry
	config   Config
}

func init() {
	Factories["interrupts"] = NewInterruptsCollector
}

// Takes a config struct and prometheus registry and returns a new Collector exposing
// interrupts stats
func NewInterruptsCollector(config Config, registry prometheus.Registry) (Collector, error) {
	c := interruptsCollector{
		config:   config,
		registry: registry,
	}
	registry.Register(
		"node_interrupts",
		"Interrupt details from /proc/interrupts",
		prometheus.NilLabels,
		interruptsMetric,
	)
	return &c, nil
}

func (c *interruptsCollector) Update() (updates int, err error) {
	interrupts, err := getInterrupts()
	if err != nil {
		return updates, fmt.Errorf("Couldn't get interrupts: %s", err)
	}
	for name, interrupt := range interrupts {
		for cpuNo, value := range interrupt.values {
			updates++
			fv, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return updates, fmt.Errorf("Invalid value %s in interrupts: %s", value, err)
			}
			labels := map[string]string{
				"CPU":     strconv.Itoa(cpuNo),
				"type":    name,
				"info":    interrupt.info,
				"devices": interrupt.devices,
			}
			interruptsMetric.Set(labels, fv)
		}
	}
	return updates, err
}

type interrupt struct {
	info    string
	devices string
	values  []string
}

func getInterrupts() (map[string]interrupt, error) {
	file, err := os.Open(procInterrupts)
	if err != nil {
		return nil, err
	}
	return parseInterrupts(file)
}

func parseInterrupts(r io.ReadCloser) (map[string]interrupt, error) {
	defer r.Close()
	interrupts := map[string]interrupt{}
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return nil, fmt.Errorf("%s empty", procInterrupts)
	}
	cpuNum := len(strings.Fields(string(scanner.Text()))) // one header per cpu

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(string(line))
		if len(parts) < cpuNum+2 { // irq + one column per cpu + details,
			continue // we ignore ERR and MIS for now
		}
		intName := parts[0][:len(parts[0])-1] // remove trailing :
		intr := interrupt{
			values: parts[1:cpuNum],
		}

		if _, err := strconv.Atoi(intName); err == nil { // numeral interrupt
			intr.info = parts[cpuNum+1]
			intr.devices = strings.Join(parts[cpuNum+2:], " ")
		} else {
			intr.info = strings.Join(parts[cpuNum+1:], " ")
		}
		interrupts[intName] = intr
	}
	return interrupts, nil
}