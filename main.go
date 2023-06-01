package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type MemoryLog struct {
	Timestamp  string  `yaml:"timestamp"`
	MemoryUsed float64 `yaml:"memoryUsed"`
}

func Logger() {
	// Maakt de file aan
	file, err := os.OpenFile("memory_logs.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Stuur de output naar de file
	log.SetOutput(file)

	// Maakt er een YAML file van
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})

	// Log de start van de monitor
	log.Info("Memory monitor started")
}

func monitorMemory(stopCh chan bool) {
	// Blijf monitoren tot er een stop signaal komt
	for {
		select {
		case <-stopCh:
			log.Info("Memory monitor stopped")
			return
		default:
			vmStat, err := mem.VirtualMemory()
			if err != nil {
				log.Errorf("Failed to get memory information: %v", err)
				return
			}

			totalUsed := float64(vmStat.Used) / 1024 / 1024 / 1024

			// Maakt een nieuwen log entry aan
			logEntry := MemoryLog{
				Timestamp:  time.Now().Format(time.RFC3339),
				MemoryUsed: totalUsed,
			}

			// Convert log entry to YAML
			logData, err := yaml.Marshal(logEntry)
			if err != nil {
				log.Errorf("Failed to convert log entry to YAML: %v", err)
				return
			}

			// Log het memory gebruik
			log.Info(string(logData))

			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	Logger()

	stopCh := make(chan bool)

	go monitorMemory(stopCh)

	// Maakt een channel aan voor de interrupt signalen
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)

	// Wacht op een interrupt signaal
	go func() {

		<-interruptCh
		stopCh <- true
	}()

	// Wacht op een enter om te stoppen
	fmt.Println("Press Enter to stop monitoring...")
	bufio.NewReader(os.Stdin).ReadString('\n')
	stopCh <- true

	// Blijf wachten tot de monitor stopt
	time.Sleep(1 * time.Second)
}
