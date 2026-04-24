package domain

import (
	"errors"
	"time"
)

var ErrInstallationStateNotFound = errors.New("installation state not found")
var ErrClusterConfigNotFound = errors.New("cluster config not found")
var ErrInvalidClusterConfig = errors.New("invalid cluster config")
var ErrTaskNotFound = errors.New("task not found")
var ErrInstallAlreadyInProgress = errors.New("install already in progress")
var ErrInstallNotRequired = errors.New("install not required")

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

type InstallationState struct {
	ManagedRoot         string
	SteamCMDInstalledAt *time.Time
	DSTInstalledAt      *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ClusterConfig struct {
	ClusterName        string
	ClusterDescription string
	GameMode           string
	MaxPlayers         int
	Language           string
	PVP                bool
	PauseWhenEmpty     bool
	Shards             []ShardConfig
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ShardConfig struct {
	Name    ShardName
	Enabled bool
}

type TaskID string

type TaskType string

const (
	TaskTypeInstallSteamCMD TaskType = "install_steamcmd"
	TaskTypeInstallDST      TaskType = "install_dst"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID         TaskID
	Type       TaskType
	Status     TaskStatus
	Detail     string
	Error      string
	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type InstallPlan struct {
	Steps []InstallStep
}

type InstallStep struct {
	Type        TaskType
	Description string
}
