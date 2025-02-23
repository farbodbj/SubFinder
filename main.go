package main

import (
	"ConfigProbe/pkg/v2rayprobe"
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/web/render"
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	filePath := flag.String("file", "", "Path to file containing subscription links (required)")
	testMethod := flag.String("method", "", "Test method [ping|speed|complete] (required)")
	flag.Parse()

	if *filePath == "" {
		slog.Error("Missing required flag", "flag", "-file")
		flag.Usage()
		os.Exit(1)
	}

	if *testMethod == "" {
		slog.Error("Missing required flag", "flag", "-method")
		flag.Usage()
		os.Exit(1)
	}

	validMethods := map[string]bool{"ping": true, "speed": true, "complete": true}
	if !validMethods[*testMethod] {
		slog.Error("Invalid test method", "valid_methods", validMethods, "received", *testMethod)
		os.Exit(1)
	}

	links, err := readSubLinks(filePath)
	if err != nil {
		slog.Error("Failed reading subscription links", "error", err)
		os.Exit(1)
	}

	v2rayTest := v2rayprobe.NewV2rayProbe(
		v2rayprobe.ConcurrencyOpt(v2rayprobe.AUTO),
		v2rayprobe.OutputMode(v2rayprobe.TEXT_OUTPUT),
	)

	countOk := testLinks(v2rayTest, links, *testMethod)
	slog.Info("Test completed", "working_nodes", countOk, "total_nodes", len(links))
}

func testLinks(v2rayTest v2rayprobe.V2rayProbe, links []string, method string) int {
	timeout := 10 * time.Second
	countOk := 0

	for _, link := range links {
		var nodes render.Nodes
		var err error

		switch method {
		case "ping":
			nodes, err = v2rayTest.TestV2RayPing(link, true, timeout)
		case "speed":
			nodes, err = v2rayTest.TestV2RaySpeed(link, true, timeout)
		case "complete":
			nodes, err = v2rayTest.TestV2RayComplete(link, true, timeout)
		default:
			slog.Error("Unhandled test method", "method", method)
			continue
		}

		if err != nil {
			slog.Warn("Test failed", "link", link, "error", err)
			continue
		}

		for _, node := range nodes {
			if node.IsOk {
				slog.Info("Working node", "link", node.Link)
				countOk++
			}
		}
	}

	return countOk
}

func readSubLinks(filePath *string) ([]string, error) {
	fmt.Println("opening file : ", *filePath)
	file, err := os.Open(*filePath)
	if err != nil {
		slog.Error("Error opening file", "err", err)
		return nil, err
	}
	defer file.Close()

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
