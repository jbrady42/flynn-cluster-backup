package main

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	configFlynn()
	watchBackup()
}

func configFlynn() {
	log.Println("Configure Flynn")
	url := os.Getenv("FLYNN_URL")
	pin := os.Getenv("FLYNN_CLUSTER_PIN")
	token := os.Getenv("FLYNN_TOKEN")

	if url == "" || pin == "" || token == "" {
		log.Fatalln("Flynn params missing")
	}

	log.Printf("Adding cluster %s", url)

	args := []string{"cluster", "add", "-p", pin, "default", url, token}
	cmd := exec.Command("flynn", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatalError("Unable to configure flynn", err)
	}
}

func watchBackup() {
	log.Println("Start backup worker")
	for true {
		startHour := currnetHour()
		runBackup()
		// Wait for next hour
		for tmp := startHour; tmp == startHour; tmp = currnetHour() {
			time.Sleep(1 * time.Minute)
		}
	}
}

func currnetHour() int {
	return time.Now().Hour()
}

func runBackup() {
	log.Println("Running backup")
	cmd := exec.Command("backup_cluster")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatalError("Unable to backup flynn", err)
	}
}

func fatalError(msg string, err error) {
	log.Printf("ERROR: %s", msg)
	log.Fatal(err)
}
