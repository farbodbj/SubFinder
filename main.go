package main

import (
	"bufio"
	"conf/pkg/v2rayprobe"
	"flag"
	"log/slog"
	"os"
	"time"
)

func main() {
	// Parse the file path from command-line arguments
	filePath := flag.String("file", "", "Path to the file containing subscription links")
	flag.Parse()

	if *filePath == "" {
		slog.Error("File path not provided")
		return
	}

	// Open the file
	file, err := os.Open(*filePath)
	if err != nil {
		slog.Error("Error opening file", "err", err)
		return
	}
	defer file.Close()

	// Read the file line by line
	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		links = append(links, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Error reading file", "err", err)
		return
	}

	// Initialize v2rayTest with desired options
	v2rayTest := v2rayprobe.NewV2rayProbe(
		v2rayprobe.ConcurrencyOpt(v2rayprobe.AUTO),
		v2rayprobe.OutputMode(v2rayprobe.TEXT_OUTPUT),
	)

	timeout := 10 * time.Second
	countOk := 0

	// Test each link
	for _, link := range links {
		nodes, _ := v2rayTest.TestV2RaySpeed(link, true, timeout)
		for _, node := range nodes {
			if node.IsOk {
				slog.Info("working nodes", "node", node.Link)
				countOk++
			}
		}
	}

	slog.Info("working nodes", "countOk", countOk)
}
