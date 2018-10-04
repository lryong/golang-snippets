package main

import "fmt"

// Snippets from https://github.com/tuna/tunasync

// Use uint8 as small capacity int.
// Use iota to define some same variable type.
type SyncStatus uint8

const (
	None SyncStatus = iota
	Failed
	Success
	Syncing
	PreSyncing
	Paused
	Disabled
)

func (s SyncStatus) String() string {
	switch s {
	case None:
		return "none"
	case Failed:
		return "failed"
	case Success:
		return "success"
	case Syncing:
		return "syncing"
	case PreSyncing:
		return "pre-syncing"
	case Paused:
		return "paused"
	case Disabled:
		return "disabled"
	default:
		return ""
	}
}

func (s *SyncStatus) UnmarshalJSON(v []byte) error {
	sv := string(v)
	switch sv {
	case `"none"`:
		*s = None
	case `"failed"`:
		*s = Failed
	case `"success"`:
		*s = Success
	case `"syncing"`:
		*s = Syncing
	case `"pre-syncing"`:
		*s = PreSyncing
	case `"paused"`:
		*s = Paused
	case `"disable"`:
		*s = Disabled
	default:
		return fmt.Errorf("Invalid status value: %s", string(v))
	}
}
