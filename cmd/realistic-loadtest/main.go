package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Duration   time.Duration
	StatusCode int
	Success    bool
	Error      error
	Endpoint   string
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
	ErrorsByType    map[string]int
}

const (
	baseURL         = "http://localhost:8080"
	targetRPS       = 5
	testDuration    = 60 * time.Second
	sliResponseTime = 300 * time.Millisecond
	sliSuccessRate  = 99.9
)

var (
	prCounter   atomic.Int64
	teamCounter atomic.Int64
)

func main() {
	fmt.Println("=== PR Reviewer Service Realistic Load Test ===")
	fmt.Printf("Target RPS: %d\n", targetRPS)
	fmt.Printf("Test Duration: %v\n", testDuration)
	fmt.Printf("SLI Response Time: %v\n", sliResponseTime)
	fmt.Printf("SLI Success Rate: %.1f%%\n\n", sliSuccessRate)

	setupTestData()

	fmt.Println("Starting realistic load test...")
	fmt.Println("Simulating real usage patterns:\n")
	results := runLoadTest()

	stats := calculateStats(results)
	printStats(stats)
	checkSLI(stats)
}

func setupTestData() {
	fmt.Println("Setting up test data...")

	for i := 1; i <= 3; i++ {
		teamPayload := map[string]any{
			"team_name": fmt.Sprintf("team-%d", i),
			"members": []map[string]any{
				{"user_id": fmt.Sprintf("u%d-1", i), "username": fmt.Sprintf("User%d-1", i), "is_active": true},
				{"user_id": fmt.Sprintf("u%d-2", i), "username": fmt.Sprintf("User%d-2", i), "is_active": true},
				{"user_id": fmt.Sprintf("u%d-3", i), "username": fmt.Sprintf("User%d-3", i), "is_active": true},
			},
		}

		makeRequest("POST", "/team/add", teamPayload)
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

	for {
		select {
		case <-timeout:
			return results
		case <-ticker.C:
			go func() {
				result := executeRealisticRequest()
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}()
		}
	}
}

func executeRealisticRequest() Result {
	operations := []func() Result{
		createPROperation,
		getTeamOperation,
		getUserReviewsOperation,
		mergePROperation,
	}

	weights := []int{40, 30, 20, 10}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	selectedOp := operations[0]

	for i, w := range weights {
		cumulative += w
		if r < cumulative {
			selectedOp = operations[i]
			break
		}
	}

	return selectedOp()
}

func createPROperation() Result {
	prNum := prCounter.Add(1)
	teamNum := (prNum % 3) + 1
	authorNum := (prNum % 3) + 1

	body := map[string]any{
		"pull_request_id":   fmt.Sprintf("pr-%d", prNum),
		"pull_request_name": fmt.Sprintf("Feature PR %d", prNum),
		"author_id":         fmt.Sprintf("u%d-%d", teamNum, authorNum),
	}

	return executeRequest("POST", "/pullRequest/create", body, "createPR")
}

func getTeamOperation() Result {
	teamNum := (rand.Intn(3) + 1)
	path := fmt.Sprintf("/team/get?team_name=team-%d", teamNum)
	return executeRequest("GET", path, nil, "getTeam")
}

func getUserReviewsOperation() Result {
	teamNum := (rand.Intn(3) + 1)
	userNum := (rand.Intn(3) + 1)
	path := fmt.Sprintf("/users/getReview?user_id=u%d-%d", teamNum, userNum)
	return executeRequest("GET", path, nil, "getUserReviews")
}

func mergePROperation() Result {
	currentPRCount := prCounter.Load()
	if currentPRCount < 10 {
		return getTeamOperation()
	}

	prNum := rand.Int63n(currentPRCount-5) + 1

	body := map[string]any{
		"pull_request_id": fmt.Sprintf("pr-%d", prNum),
	}

	return executeRequest("POST", "/pullRequest/merge", body, "mergePR")
}

func executeRequest(method, path string, body any, endpoint string) Result {
	start := time.Now()
	resp, err := makeRequest(method, path, body)
	duration := time.Since(start)

	result := Result{
		Duration: duration,
		Error:    err,
		Endpoint: endpoint,
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
	errorsByType := make(map[string]int)

	for _, r := range results {
		durations = append(durations, r.Duration)
		totalDuration += r.Duration
		if r.Success {
			successCount++
		} else {
			if r.Error != nil {
				errorsByType["network_error"]++
			} else {
				errorsByType[fmt.Sprintf("http_%d", r.StatusCode)]++
			}
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
		ErrorsByType:    errorsByType,
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

	if len(stats.ErrorsByType) > 0 {
		fmt.Println("\nError Breakdown:")
		for errType, count := range stats.ErrorsByType {
			fmt.Printf("  %s: %d\n", errType, count)
		}
	}

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
