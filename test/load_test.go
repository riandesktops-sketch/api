package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	concurrent = flag.Int("concurrent", 10, "Number of concurrent requests")
	requests   = flag.Int("requests", 100, "Total number of requests")
	url        = flag.String("url", "http://localhost:8084/api/v1/ai/chat", "AI service URL")
)

type ChatRequest struct {
	ZodiacSign  string `json:"zodiac_sign"`
	UserMessage string `json:"user_message"`
}

type ChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Response string `json:"response"`
	} `json:"data"`
}

type Stats struct {
	totalRequests   int32
	successRequests int32
	failedRequests  int32
	queueFullErrors int32
	totalLatency    int64 // in milliseconds
	minLatency      int64
	maxLatency      int64
}

func main() {
	flag.Parse()

	log.Printf("🚀 Starting load test...")
	log.Printf("📊 Config: %d concurrent, %d total requests", *concurrent, *requests)
	log.Printf("🎯 Target: %s", *url)

	stats := &Stats{
		minLatency: 999999,
	}

	startTime := time.Now()

	// Create semaphore for concurrency control
	sem := make(chan struct{}, *concurrent)
	var wg sync.WaitGroup

	// Send requests
	for i := 0; i < *requests; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(reqNum int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			sendRequest(reqNum, stats)
		}(i + 1)

		// Small delay to avoid overwhelming the system
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all requests to complete
	wg.Wait()
	totalDuration := time.Since(startTime)

	// Print results
	printStats(stats, totalDuration)
}

func sendRequest(reqNum int, stats *Stats) {
	atomic.AddInt32(&stats.totalRequests, 1)

	reqBody := ChatRequest{
		ZodiacSign:  "Gemini",
		UserMessage: fmt.Sprintf("Test message #%d - Halo, apa kabar?", reqNum),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("❌ Request #%d: Failed to marshal JSON: %v", reqNum, err)
		atomic.AddInt32(&stats.failedRequests, 1)
		return
	}

	startTime := time.Now()

	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Request #%d: HTTP error: %v", reqNum, err)
		atomic.AddInt32(&stats.failedRequests, 1)
		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	// Update latency stats
	atomic.AddInt64(&stats.totalLatency, latency)
	for {
		min := atomic.LoadInt64(&stats.minLatency)
		if latency >= min || atomic.CompareAndSwapInt64(&stats.minLatency, min, latency) {
			break
		}
	}
	for {
		max := atomic.LoadInt64(&stats.maxLatency)
		if latency <= max || atomic.CompareAndSwapInt64(&stats.maxLatency, max, latency) {
			break
		}
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		log.Printf("❌ Request #%d: Failed to decode response: %v", reqNum, err)
		atomic.AddInt32(&stats.failedRequests, 1)
		return
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		log.Printf("⚠️ Request #%d: Queue full (429) - %dms", reqNum, latency)
		atomic.AddInt32(&stats.queueFullErrors, 1)
		atomic.AddInt32(&stats.failedRequests, 1)
		return
	}

	if resp.StatusCode != http.StatusOK || !chatResp.Success {
		log.Printf("❌ Request #%d: Failed (status %d) - %dms", reqNum, resp.StatusCode, latency)
		atomic.AddInt32(&stats.failedRequests, 1)
		return
	}

	atomic.AddInt32(&stats.successRequests, 1)
	log.Printf("✅ Request #%d: Success - %dms - Response: %.50s...", 
		reqNum, latency, chatResp.Data.Response)
}

func printStats(stats *Stats, duration time.Duration) {
	total := atomic.LoadInt32(&stats.totalRequests)
	success := atomic.LoadInt32(&stats.successRequests)
	failed := atomic.LoadInt32(&stats.failedRequests)
	queueFull := atomic.LoadInt32(&stats.queueFullErrors)
	totalLatency := atomic.LoadInt64(&stats.totalLatency)
	minLatency := atomic.LoadInt64(&stats.minLatency)
	maxLatency := atomic.LoadInt64(&stats.maxLatency)

	avgLatency := int64(0)
	if total > 0 {
		avgLatency = totalLatency / int64(total)
	}

	successRate := float64(success) / float64(total) * 100
	throughput := float64(total) / duration.Seconds()

	fmt.Println("\n" + "="*60)
	fmt.Println("📊 LOAD TEST RESULTS")
	fmt.Println("="*60)
	fmt.Printf("⏱️  Total Duration:     %v\n", duration)
	fmt.Printf("📨 Total Requests:     %d\n", total)
	fmt.Printf("✅ Success:            %d (%.2f%%)\n", success, successRate)
	fmt.Printf("❌ Failed:             %d\n", failed)
	fmt.Printf("🚫 Queue Full (429):   %d\n", queueFull)
	fmt.Println("-"*60)
	fmt.Printf("⚡ Throughput:         %.2f req/s\n", throughput)
	fmt.Printf("⏱️  Avg Latency:        %dms\n", avgLatency)
	fmt.Printf("⏱️  Min Latency:        %dms\n", minLatency)
	fmt.Printf("⏱️  Max Latency:        %dms\n", maxLatency)
	fmt.Println("="*60)

	if successRate >= 95 {
		fmt.Println("🎉 EXCELLENT! Success rate >= 95%")
	} else if successRate >= 80 {
		fmt.Println("✅ GOOD! Success rate >= 80%")
	} else if successRate >= 50 {
		fmt.Println("⚠️  NEEDS IMPROVEMENT! Success rate < 80%")
	} else {
		fmt.Println("❌ POOR! Success rate < 50%")
	}
}
