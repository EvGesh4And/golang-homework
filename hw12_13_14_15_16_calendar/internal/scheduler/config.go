package scheduler

import "time"

type NotificationsConf struct {
	Tick     time.Duration `toml:"tick"`
	EventTTL time.Duration `toml:"event_ttl"`
}
