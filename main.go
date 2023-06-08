package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type SystemStats struct {
	Timestamp  string  `yaml:"timestamp"`
	MemoryUsed float64 `yaml:"memoryUsed"`
	CPUUsed    float64 `yaml:"cpuUsed"`
}

func Logger() {
	// Maakt de file aan
	file, err := os.OpenFile("system_logs.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
	log.Info("System monitor started")
}

func monitorMemory(stopCh chan bool) {
	// Blijf monitoren tot er een stop signaal komt
	for {
		select {
		case <-stopCh:
			log.Info("Memory monitor stopped")
			return
		default:
			// Gebruik gopsutil om het geheugeninformatie op te halen
			vmStat, err := mem.VirtualMemory()
			if err != nil {
				log.Errorf("Failed to get memory information: %v", err)
				return
			}

			totalUsed := float64(vmStat.Used) / 1024 / 1024 / 1024

			// Maakt een nieuwe log entry aan
			logEntry := SystemStats{
				Timestamp:  time.Now().Format(time.RFC3339),
				MemoryUsed: totalUsed,
			}

			// Convert log entry to YAML
			logData, err := yaml.Marshal(logEntry)
			if err != nil {
				log.Errorf("Failed to convert log entry to YAML: %v", err)
				return
			}

			// Log het geheugengebruik
			log.Info(string(logData))

			time.Sleep(1 * time.Second)
		}
	}
}

func monitorCPU(stopCh chan bool) {
	// Blijf monitoren tot er een stop signaal komt
	for {
		select {
		case <-stopCh:
			log.Info("CPU monitor stopped")
			return
		default:
			// Gebruik gopsutil om de CPU-gebruiksinformatie op te halen
			cpuStat, err := cpu.Percent(0, false)
			if err != nil {
				log.Errorf("Failed to get CPU information: %v", err)
				return
			}

			cpuUsed := cpuStat[0]

			// Maakt een nieuwe log entry aan
			logEntry := SystemStats{
				Timestamp: time.Now().Format(time.RFC3339),
				CPUUsed:   cpuUsed,
			}

			// Convert log entry to YAML
			logData, err := yaml.Marshal(logEntry)
			if err != nil {
				log.Errorf("Failed to convert log entry to YAML: %v", err)
				return
			}

			// Log het CPU-gebruik
			log.Info(string(logData))

			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	Logger()

	// Vraag aan de gebruiker welke monitor ze willen uitvoeren
	fmt.Println("Which monitor do you want to run? Enter 'cpu' or 'memory':")
	reader := bufio.NewReader(os.Stdin)
	monitorChoice, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read user input: %v", err)
	}
	monitorChoice = monitorChoice[:len(monitorChoice)-2] // Verwijder de newline-karakter en het carriage return-karakter van de input

	stopCh := make(chan bool)

	if monitorChoice == "cpu" {
		go monitorCPU(stopCh)
	} else if monitorChoice == "memory" {
		go monitorMemory(stopCh)
	} else {
		log.Fatalf("Invalid monitor choice. Exiting...")
	}

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
