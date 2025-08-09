package db

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func MakeInfluxPoint(server string, online, max int64, ping time.Duration) *write.Point {
	tags := map[string]string{"server": server}
	fields := map[string]any{
		"online":  online,
		"max":     max,
		"ping_ms": ping.Milliseconds(),
	}

	return influxdb2.NewPoint(
		"mcserver_status",
		tags,
		fields,
		time.Now(),
	)
}
