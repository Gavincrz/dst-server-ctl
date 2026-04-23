package domain

import "time"

type ShardName string

const (
	ShardMaster ShardName = "Master"
	ShardCaves  ShardName = "Caves"
)

type ManagedLayout struct {
	Root     string
	SteamCMD string
	DST      string
	Clusters string
	Logs     string
	State    string
}

type ServerStatus string

const (
	ServerStatusUnknown ServerStatus = "unknown"
	ServerStatusStopped ServerStatus = "stopped"
	ServerStatusRunning ServerStatus = "running"
)

type Status struct {
	Version   string       `json:"version"`
	Status    ServerStatus `json:"status"`
	StartedAt *time.Time   `json:"startedAt,omitempty"`
}

