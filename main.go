package main

import (
	"ConfigProbe/pkg/v2rayprobe"
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	links, err := readSubLinks(getFilePathFromArgs())
	if err != nil {
		slog.Error("Reading subscription links", "Error occured when reading subscription links", err)
	}

	// Initialize v2rayTest with desired options
	v2rayTest := v2rayprobe.NewV2rayProbe(
		v2rayprobe.ConcurrencyOpt(v2rayprobe.AUTO),
		v2rayprobe.OutputMode(v2rayprobe.TEXT_OUTPUT),
	)

	countOk := testLinks(v2rayTest, links)
	if countOk == 0 {
		// re-perform the test
		countOk = testLinks(v2rayTest, links)
	}

	slog.Info("working nodes", "countOk", countOk)
}

func testLinks(v2rayTest v2rayprobe.V2rayProbe, links []string) int {
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

	return countOk
}

func getFilePathFromArgs() *string {
	// Parse the file path from command-line arguments
	filePath := flag.String("file", "", "Path to the file containing subscription links")
	flag.Parse()

	if *filePath == "" {
		slog.Error("File path not provided")
		return nil
	}
	return filePath
}

func readSubLinks(filePath *string) ([]string, error) {
	// Open the file
	fmt.Println("opening file : ", *filePath)
	file, err := os.Open(*filePath)
	if err != nil {
		slog.Error("Error opening file", "err", err)
		return nil, err
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
		return nil, err
	}

	return links, nil
}
