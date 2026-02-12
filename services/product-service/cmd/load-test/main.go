package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Metrics struct {
	TotalRequests int64
	SuccessCount  int64
	FailureCount  int64
	MinLatency    int64
	MaxLatency    int64
	TotalLatency  int64
	mu            sync.RWMutex
}

func (m *Metrics) RecordSuccess(latency int64) {
	atomic.AddInt64(&m.TotalRequests, 1)
	atomic.AddInt64(&m.SuccessCount, 1)
	atomic.AddInt64(&m.TotalLatency, latency)

	m.mu.Lock()
	if m.MinLatency == 0 || latency < m.MinLatency {
		m.MinLatency = latency
	}
	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}
	m.mu.Unlock()
}

func (m *Metrics) RecordFailure() {
	atomic.AddInt64(&m.TotalRequests, 1)
	atomic.AddInt64(&m.FailureCount, 1)
}

func (m *Metrics) PrintReport() {
	total := atomic.LoadInt64(&m.TotalRequests)
	success := atomic.LoadInt64(&m.SuccessCount)
	failure := atomic.LoadInt64(&m.FailureCount)
	totalLatency := atomic.LoadInt64(&m.TotalLatency)

	m.mu.RLock()
	min := m.MinLatency
	max := m.MaxLatency
	m.mu.RUnlock()

	successRate := float64(success) / float64(total) * 100
	avgLatency := float64(totalLatency) / float64(success)

	fmt.Printf("\n=== Load Test Results ===\n")
	fmt.Printf("Total Requests:    %d\n", total)
	fmt.Printf("Successful:        %d\n", success)
	fmt.Printf("Failed:            %d\n", failure)
	fmt.Printf("Success Rate:      %.2f%%\n", successRate)
	fmt.Printf("\nLatency (ms):\n")
	fmt.Printf("  Min:            %d\n", min)
	fmt.Printf("  Max:            %d\n", max)
	fmt.Printf("  Average:         %.2f\n", avgLatency)
}

func runConcurrentTest(client productpb.ProductServiceClient, productID string, concurrentRequests int) *Metrics {
	metrics := &Metrics{}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5000)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			req := &productpb.GetProductRequest{ProductId: productID}

			latencyStart := time.Now()
			_, err := client.GetProduct(context.Background(), req)
			latency := time.Since(latencyStart).Milliseconds()

			if err != nil {
				metrics.RecordFailure()
			} else {
				metrics.RecordSuccess(latency)
			}
		}()
	}

	wg.Wait()
	return metrics
}

func runCacheBenchmark(client productpb.ProductServiceClient) {
	fmt.Println("\n=== Cache Performance Benchmark ===")
	fmt.Println("This will compare cache miss vs cache hit performance")

	fmt.Println("Step 1: Cache Miss Test (Cold Start)")
	fmt.Println("Fetching product from database...")
	productID := fmt.Sprintf("%d", time.Now().UnixNano())
	startTime := time.Now()

	metricsMiss := runConcurrentTest(client, productID, 1000)
	durationMiss := time.Since(startTime)

	metricsMiss.PrintReport()
	fmt.Printf("Duration:          %v\n", durationMiss)
	fmt.Printf("Requests/Second:    %.2f\n", float64(1000)/durationMiss.Seconds())

	fmt.Println("\nStep 2: Cache Hit Test (Warm Cache)")
	fmt.Println("Warming up cache with first request...")
	_, _ = client.GetProduct(context.Background(), &productpb.GetProductRequest{ProductId: productID})
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Running concurrent reads from cache...")
	startTime = time.Now()

	metricsHit := runConcurrentTest(client, productID, 1000)
	durationHit := time.Since(startTime)

	metricsHit.PrintReport()
	fmt.Printf("Duration:          %v\n", durationHit)
	fmt.Printf("Requests/Second:    %.2f\n", float64(1000)/durationHit.Seconds())

	fmt.Println("\n=== Comparison ===")
	improvement := float64(durationMiss.Milliseconds()) / float64(durationHit.Milliseconds())
	fmt.Printf("Cache is %.2fx faster than database\n", improvement)
}

func main() {
	conn, err := grpc.NewClient("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect: %v", err))
	}
	defer func() { _ = conn.Close() }()

	client := productpb.NewProductServiceClient(conn)

	if len(os.Args) > 1 && os.Args[1] == "benchmark" {
		runCacheBenchmark(client)
		return
	}

	fmt.Println("=== Product Service Load Test ===")
	fmt.Println("Target: 10,000 concurrent read requests")
	fmt.Println("Service: localhost:9091")

	const concurrentRequests = 10000
	productID := os.Args[1]
	if productID == "" {
		productID = "550e8400-e29b-41d4-a716-446655440000"
	}

	startTime := time.Now()

	metrics := runConcurrentTest(client, productID, concurrentRequests)

	totalDuration := time.Since(startTime)

	metrics.PrintReport()
	fmt.Printf("\nTotal Duration:     %v\n", totalDuration)
	fmt.Printf("Requests/Second:    %.2f\n", float64(concurrentRequests)/totalDuration.Seconds())
	fmt.Printf("\nPerformance Targets:")
	fmt.Printf("  Success Rate:     > 95%%\n")
	fmt.Printf("  Avg Latency:     < 100ms\n")
	fmt.Printf("  Req/Sec:         > 1000\n")
}
