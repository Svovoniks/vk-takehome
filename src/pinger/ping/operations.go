package ping

import (
	"time"

	"github.com/prometheus-community/pro-bing"
)

func DoPing(ip string) (float64, error) {
	pinger, err := probing.NewPinger(ip)
	if err != nil {
		return -1, err
	}

	pinger.Count = 1
	pinger.Timeout = time.Second

	err = pinger.Run()
	if err != nil {
		return -1, err
	}

	stats := pinger.Statistics()
	return float64(stats.MaxRtt.Microseconds()) / 1000, nil
}
