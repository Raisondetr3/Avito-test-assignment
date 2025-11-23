package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Result struct {
	Duration   time.Duration
	StatusCode int
	Success    bool
	Error      error
}

type Stats struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	MinDuration     time.Duration
	MaxDuration     time.Duration
	AvgDuration     time.Duration
	P50Duration     time.Duration
	P95Duration     time.Duration
	P99Duration     time.Duration
	SuccessRate     float64
	ActualRPS       float64
}

const (
	baseURL     = "http://localhost:8080"
	targetRPS   = 5
	testDuration = 60 * time.Second
	sliResponseTime = 300 * time.Millisecond
	sliSuccessRate  = 99.9
)

func main() {
	fmt.Println("=== PR Reviewer Service Load Test ===")
	fmt.Printf("Target RPS: %d\n", targetRPS)
	fmt.Printf("Test Duration: %v\n", testDuration)
	fmt.Printf("SLI Response Time: %v\n", sliResponseTime)
	fmt.Printf("SLI Success Rate: %.1f%%\n\n", sliSuccessRate)

	setupTestData()

	fmt.Println("Starting load test...")
	results := runLoadTest()

	stats := calculateStats(results)
	printStats(stats)
	checkSLI(stats)
}

func setupTestData() {
	fmt.Println("Setting up test data...")

	teamPayload := map[string]any{
		"team_name": "loadtest-team",
		"members": []map[string]any{
			{"user_id": "lt-u1", "username": "LoadUser1", "is_active": true},
			{"user_id": "lt-u2", "username": "LoadUser2", "is_active": true},
			{"user_id": "lt-u3", "username": "LoadUser3", "is_active": true},
		},
	}

	resp, err := makeRequest("POST", "/team/add", teamPayload)
	if err != nil {
		fmt.Printf("Warning: Failed to create team: %v\n", err)
	} else if resp.StatusCode != 201 && resp.StatusCode != 400 {
		fmt.Printf("Warning: Unexpected status code when creating team: %d\n", resp.StatusCode)
	}

	fmt.Println("Test data setup complete\n")
}

func runLoadTest() []Result {
	var results []Result
	var mu sync.Mutex

	requestInterval := time.Second / time.Duration(targetRPS)
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	timeout := time.After(testDuration)
	requestNum := 0

	for {
		select {
		case <-timeout:
			return results
		case <-ticker.C:
			requestNum++
			go func(num int) {
				result := executeRequest(num)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(requestNum)
		}
	}
}

func executeRequest(num int) Result {
	endpoints := []struct {
		method string
		path   string
		body   any
	}{
		{"GET", "/team/get?team_name=loadtest-team", nil},
		{"GET", "/users/getReview?user_id=lt-u1", nil},
		{"POST", "/pullRequest/create", map[string]any{
			"pull_request_id":   fmt.Sprintf("lt-pr-%d", num),
			"pull_request_name": fmt.Sprintf("Load Test PR %d", num),
			"author_id":         "lt-u1",
		}},
		{"POST", "/pullRequest/merge", map[string]any{
			"pull_request_id": fmt.Sprintf("lt-pr-%d", num),
		}},
	}

	endpoint := endpoints[num%len(endpoints)]

	start := time.Now()
	resp, err := makeRequest(endpoint.method, endpoint.path, endpoint.body)
	duration := time.Since(start)

	result := Result{
		Duration: duration,
		Error:    err,
	}

	if err != nil {
		result.Success = false
		result.StatusCode = 0
	} else {
		result.StatusCode = resp.StatusCode
		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	return result
}

func makeRequest(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return client.Do(req)
}

func calculateStats(results []Result) Stats {
	if len(results) == 0 {
		return Stats{}
	}

	var durations []time.Duration
	successCount := 0
	var totalDuration time.Duration

	for _, r := range results {
		durations = append(durations, r.Duration)
		totalDuration += r.Duration
		if r.Success {
			successCount++
		}
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	stats := Stats{
		TotalRequests:   len(results),
		SuccessRequests: successCount,
		FailedRequests:  len(results) - successCount,
		MinDuration:     durations[0],
		MaxDuration:     durations[len(durations)-1],
		AvgDuration:     totalDuration / time.Duration(len(results)),
		P50Duration:     durations[len(durations)*50/100],
		P95Duration:     durations[len(durations)*95/100],
		P99Duration:     durations[len(durations)*99/100],
		SuccessRate:     float64(successCount) / float64(len(results)) * 100,
		ActualRPS:       float64(len(results)) / testDuration.Seconds(),
	}

	return stats
}

func printStats(stats Stats) {
	fmt.Println("\n=== Load Test Results ===")
	fmt.Printf("Total Requests:    %d\n", stats.TotalRequests)
	fmt.Printf("Success:           %d\n", stats.SuccessRequests)
	fmt.Printf("Failed:            %d\n", stats.FailedRequests)
	fmt.Printf("Success Rate:      %.2f%%\n", stats.SuccessRate)
	fmt.Printf("Actual RPS:        %.2f\n", stats.ActualRPS)
	fmt.Println("\nResponse Times:")
	fmt.Printf("  Min:             %v\n", stats.MinDuration)
	fmt.Printf("  Avg:             %v\n", stats.AvgDuration)
	fmt.Printf("  P50 (median):    %v\n", stats.P50Duration)
	fmt.Printf("  P95:             %v\n", stats.P95Duration)
	fmt.Printf("  P99:             %v\n", stats.P99Duration)
	fmt.Printf("  Max:             %v\n", stats.MaxDuration)
}

func checkSLI(stats Stats) {
	fmt.Println("\n=== SLI Compliance Check ===")

	responseTimeMet := stats.P95Duration <= sliResponseTime
	successRateMet := stats.SuccessRate >= sliSuccessRate

	fmt.Printf("Response Time SLI (P95 <= %v): ", sliResponseTime)
	if responseTimeMet {
		fmt.Printf("✓ PASS (P95: %v)\n", stats.P95Duration)
	} else {
		fmt.Printf("✗ FAIL (P95: %v)\n", stats.P95Duration)
	}

	fmt.Printf("Success Rate SLI (>= %.1f%%): ", sliSuccessRate)
	if successRateMet {
		fmt.Printf("✓ PASS (%.2f%%)\n", stats.SuccessRate)
	} else {
		fmt.Printf("✗ FAIL (%.2f%%)\n", stats.SuccessRate)
	}

	fmt.Println()
	if responseTimeMet && successRateMet {
		fmt.Println("🎉 All SLI requirements are MET!")
	} else {
		fmt.Println("⚠️  Some SLI requirements are NOT MET")
	}
}
