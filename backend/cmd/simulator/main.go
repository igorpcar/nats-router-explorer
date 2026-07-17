package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

type DeviceMetric struct {
	Topic    string
	Generate func(r *rand.Rand) interface{}
}

func main() {
	natsURL := flag.String("nats", "nats://localhost:4222", "NATS server URL")
	interval := flag.Duration("interval", 1500*time.Millisecond, "Publish interval")
	flag.Parse()

	log.Printf("Connecting to NATS at %s...", *natsURL)
	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Successfully connected to NATS!")

	// initialize local random generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// define simulated metrics
	metrics := []DeviceMetric{
		{
			Topic: "iot_domain_brazil.factory1.temperature",
			Generate: func(r *rand.Rand) interface{} {
				return map[string]interface{}{
					"valor":   20.0 + r.Float64()*15.0,
					"unidade": "C",
					"status":  "ok",
				}
			},
		},
		{
			Topic: "iot_domain_brazil.factory1.humidity",
			Generate: func(r *rand.Rand) interface{} {
				return map[string]interface{}{
					"valor":   50.0 + r.Float64()*30.0,
					"unidade": "%",
					"status":  "ok",
				}
			},
		},
		{
			Topic: "iot_domain_usa.turbine3.wind_speed",
			Generate: func(r *rand.Rand) interface{} {
				return map[string]interface{}{
					"valor":     5.0 + r.Float64()*15.0,
					"unidade":   "m/s",
					"rotor_rpm": 10.0 + r.Float64()*12.0,
				}
			},
		},
		{
			Topic: "iot_domain_usa.turbine3.power_output",
			Generate: func(r *rand.Rand) interface{} {
				return map[string]interface{}{
					"valor":      500 + r.Intn(1000),
					"unidade":    "kW",
					"efficiency": 85.0 + r.Float64()*10.0,
				}
			},
		},
		{
			Topic: "iot_domain_germany.assembly.conveyor",
			Generate: func(r *rand.Rand) interface{} {
				statuses := []string{"running", "paused", "stopped"}
				status := statuses[r.Intn(len(statuses))]
				speed := 0.0
				if status == "running" {
					speed = 0.8 + r.Float64()*0.8
				}
				return map[string]interface{}{
					"status": status,
					"speed":  speed,
					"load":   30.0 + r.Float64()*60.0,
				}
			},
		},
		{
			Topic: "iot_domain_germany.assembly.robot_arm",
			Generate: func(r *rand.Rand) interface{} {
				return map[string]interface{}{
					"active":      r.Float64() > 0.15,
					"cycle_count": 14205 + r.Intn(100),
					"joint_temperatures": []float64{
						30.0 + r.Float64()*10,
						32.0 + r.Float64()*12,
						35.0 + r.Float64()*15,
					},
				}
			},
		},
		{
			Topic: "iot_domain_france.refinery.pressure",
			Generate: func(r *rand.Rand) interface{} {
				pressure := 3.5 + r.Float64()*2.0
				return map[string]interface{}{
					"valor":   pressure,
					"unidade": "bar",
					"alarm":   pressure > 5.0,
				}
			},
		},
		{
			Topic: "iot_domain_france.refinery.flow_rate",
			Generate: func(r *rand.Rand) interface{} {
				return map[string]interface{}{
					"valor":   100.0 + r.Float64()*50.0,
					"unidade": "L/min",
					"temp":    20.0 + r.Float64()*15.0,
				}
			},
		},
	}

	log.Printf("Starting IoT simulation publishing to %d topics every %v...", len(metrics), *interval)
	log.Println("Press Ctrl+C to terminate.")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// send messages immediately to populate screen quickly
	log.Println("Populating initial topics...")
	for _, m := range metrics {
		data := m.Generate(r)
		payload, err := json.Marshal(data)
		if err == nil {
			nc.Publish(m.Topic, payload)
		}
	}

	for {
		select {
		case <-ticker.C:
			// publish one random metric
			metric := metrics[r.Intn(len(metrics))]
			data := metric.Generate(r)

			payload, err := json.Marshal(data)
			if err != nil {
				log.Printf("Error serializing JSON for %s: %v", metric.Topic, err)
				continue
			}

			err = nc.Publish(metric.Topic, payload)
			if err != nil {
				log.Printf("Error publishing to topic %s: %v", metric.Topic, err)
			} else {
				log.Printf("[PUB] %s -> %s", metric.Topic, string(payload))
			}

		case <-stopChan:
			log.Println("Terminating IoT simulator...")
			return
		}
	}
}
