package ping

import "time"

type PingEvent struct {
	Ip        string    `json:"ip"`
	Ping_ms   float64   `json:"ping_ms"`
	Pinged_at time.Time `json:"pinged_at"`
}

type PingEventList = []PingEvent
