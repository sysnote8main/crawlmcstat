package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"sync"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sysnote8main/crawlmcstat/internal/config"
	"github.com/sysnote8main/crawlmcstat/internal/db"
	"github.com/sysnote8main/crawlmcstat/internal/mcsrv"
	"golang.org/x/sync/semaphore"
)

func pollServersAndWrite(
	ctx context.Context,
	servers []string,
	timeout time.Duration,
	maxConcurrent int64,
	writeAPI api.WriteAPI,
) {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(maxConcurrent)
	results := make([]string, len(servers))

	for i, addr := range servers {
		wg.Add(1)
		go func(i int, addr string) {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				results[i] = fmt.Sprintf("%s -> semaphore error: %v", addr, err)
				return
			}
			defer sem.Release(1)

			online, max, ping, err := mcsrv.GetPlayerCountWithTimeout(addr, timeout)
			if err != nil {
				return
			}
			results[i] = fmt.Sprintf("%s -> Players: %d/%d, Ping: %dms", addr, online, max, ping.Milliseconds())

			point := db.MakeInfluxPoint(addr, online, max, ping)
			writeAPI.WritePoint(point)
		}(i, addr)
	}

	wg.Wait()
	writeAPI.Flush()

	slog.Info("Crawled!")
}

func main() {
	slog.Info("Loading config...")
	const configPath = "config.yml"

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	// config
	dbCfg := cfg.DBSetting
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second

	slog.Info("Connecting database...")
	// setup database
	client := influxdb2.NewClient(dbCfg.Url, dbCfg.Token)
	defer client.Close()
	writeAPI := client.WriteAPI(dbCfg.Organization, dbCfg.Bucket)

	// create ticker
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	slog.Info("Start crawling!")
	// do loop
	ctx := context.Background()
	for {
		pollServersAndWrite(ctx, cfg.Servers, timeout, cfg.MaxConcurrent, writeAPI)
		<-ticker.C
	}
}
