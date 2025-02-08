package ping

import "time"

type PingEvent struct {
	Ip        string    `json:"ip"`
	Ping_ms   int       `json:"ping_ms"`
	Pinged_at time.Time `json:"pinged_at"`
}
