package scheduler

import "time"

// NotificationsConf stores scheduler timing configuration.
type NotificationsConf struct {
	Tick     time.Duration `toml:"tick"`
	EventTTL time.Duration `toml:"event_ttl"`
}
