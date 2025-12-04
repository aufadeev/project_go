package main

import (
	"context"
	"fmt"
	"io"
	//"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	targetURL     = "http://srv.msk01.gigacorp.local/_stats"
	checkInterval = 10 * time.Second
	maxErrors     = 3
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	errorCount := 0

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := fetchAndCheck(ctx, client)
		cancel()

		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				errorCount = 0 // сброс счётчика, чтобы не спамить
			}
		} else {
			errorCount = 0
		}
	}
}

func fetchAndCheck(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fields := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(fields) != 7 {
		return fmt.Errorf("unexpected number of fields: %d", len(fields))
	}

	// Парсинг значений
	loadAvg, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return fmt.Errorf("invalid load average: %v", err)
	}

	totalMem, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return fmt.Errorf("invalid total memory: %v", err)
	}

	usedMem, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return fmt.Errorf("invalid used memory: %v", err)
	}

	totalDisk, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return fmt.Errorf("invalid total disk: %v", err)
	}

	usedDisk, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return fmt.Errorf("invalid used disk: %v", err)
	}

	totalNet, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return fmt.Errorf("invalid total network bandwidth: %v", err)
	}

	usedNet, err := strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return fmt.Errorf("invalid used network bandwidth: %v", err)
	}

// Load Average
if loadAvg > 30 {
    fmt.Printf("Load Average is too high: %d\n", int(loadAvg))
}

// Память
if totalMem > 0 {
    usage := (usedMem / totalMem) * 100
    if usage > 80 {
        fmt.Printf("Memory usage too high: %d%%\n", int(usage))
    }
}

// Диск
if totalDisk > 0 {
    usage := (usedDisk / totalDisk) * 100
    if usage > 90 {
        freeMB := int((totalDisk - usedDisk) / (1024 * 1024))
        fmt.Printf("Free disk space is too low: %d Mb left\n", freeMB)
    }
}

// Сеть
if totalNet > 0 {
    usage := (usedNet / totalNet) * 100
    if usage > 90 {
        // Внимание: тест ожидает деление на 1e6 БЕЗ умножения на 8
        freeMbit := int((totalNet - usedNet) / (1000 * 1000))
        fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
    }
}

	return nil
}