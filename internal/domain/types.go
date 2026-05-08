package domain

import (
	"context"
	"errors"
	"time"
)

var ErrInstallationStateNotFound = errors.New("installation state not found")
var ErrUpdateStateNotFound = errors.New("update state not found")
var ErrClusterConfigNotFound = errors.New("cluster config not found")
var ErrInvalidClusterConfig = errors.New("invalid cluster config")
var ErrTaskNotFound = errors.New("task not found")
var ErrInstallAlreadyInProgress = errors.New("install already in progress")
var ErrInstallNotRequired = errors.New("install not required")
var ErrUpdateAlreadyInProgress = errors.New("update already in progress")
var ErrUpdateNotRequired = errors.New("update not required")
var ErrUpdateRequiresServerStop = errors.New("update requires server stop confirmation")
var ErrDSTNotInstalled = errors.New("dst not installed")
var ErrServerAlreadyRunning = errors.New("server already running")
var ErrServerNotRunning = errors.New("server not running")
var ErrInvalidShard = errors.New("invalid shard")

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

type LogStreamUpdate struct {
	Lines   []string
	Reset   bool
	Changed bool
}

type LogStream interface {
	Snapshot() []string
	Poll(ctx context.Context) (LogStreamUpdate, error)
	Close() error
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

type RuntimeStatus struct {
	Status          ServerStatus `json:"status"`
	Shards          []ShardState `json:"shards"`
	RestartRequired bool         `json:"restartRequired"`
	LastError       string       `json:"lastError,omitempty"`
}

type ShardState struct {
	Name    ShardName `json:"name"`
	Running bool      `json:"running"`
	PID     int       `json:"pid,omitempty"`
}

type RuntimeEventKind string

const (
	RuntimeEventStarted RuntimeEventKind = "started"
	RuntimeEventStopped RuntimeEventKind = "stopped"
	RuntimeEventExited  RuntimeEventKind = "exited"
	RuntimeEventRetried RuntimeEventKind = "retried"
)

type RuntimeEvent struct {
	ID        int64            `json:"id"`
	Shard     ShardName        `json:"shard"`
	Kind      RuntimeEventKind `json:"kind"`
	Detail    string           `json:"detail"`
	CreatedAt time.Time        `json:"createdAt"`
}

type InstallationState struct {
	ManagedRoot         string
	SteamCMDInstalledAt *time.Time
	DSTInstalledAt      *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type UpdateState struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	LastCheckedAt   *time.Time
	LastUpdatedAt   *time.Time
	LastError       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ClusterConfig struct {
	ClusterName        string
	ClusterDescription string
	ClusterPassword    string
	ClusterIntention   string
	GameMode           string
	MaxPlayers         int
	Language           string
	PVP                bool
	PauseWhenEmpty     bool
	OfflineCluster     bool
	LANOnlyCluster     bool
	TickRate           int
	ConsoleEnabled     bool
	BindIP             string
	MasterPort         int
	ClusterKey         string
	Shards             []ShardConfig
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ShardConfig struct {
	Name               ShardName
	Enabled            bool
	ServerPort         int
	MasterServerPort   int
	AuthenticationPort int
}

type TaskID string

type TaskType string

const (
	TaskTypeInstallSteamCMD TaskType = "install_steamcmd"
	TaskTypeInstallDST      TaskType = "install_dst"
	TaskTypeUpdateCheckDST  TaskType = "check_dst_update"
	TaskTypeUpdateDST       TaskType = "update_dst"
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
