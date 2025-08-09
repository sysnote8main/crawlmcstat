package mcsrv

import (
	"log/slog"
	"time"

	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/options"
)

func GetPlayerCountWithTimeout(address string, timeout time.Duration) (online, max int64, ping time.Duration, err error) {
	ip, port, err := mcutil.ParseAddress(address, 25565)
	if err != nil {
		slog.Error("Failed to parse address", slog.String("address", address), slog.Any("error", err))
		return 0, 0, 0, err
	}
	resp, err := mcutil.Status(ip, port, options.JavaStatus{Timeout: timeout})
	if err != nil {
		slog.Error("Failed to parse address", slog.String("ip", ip), slog.Int("port", int(port)), slog.Any("error", err))
		return 0, 0, 0, err
	}

	return *resp.Players.Online, *resp.Players.Max, resp.Latency, nil
}
