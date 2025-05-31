package sqlstorage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func parsePostgresInterval(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	// Пример строки: "1 day 02:03:04"
	re := regexp.MustCompile(`(?:(\d+) day[s]?)?\s*(\d+):(\d+):(\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid interval format: %q", s)
	}

	days := 0
	if matches[1] != "" {
		days, _ = strconv.Atoi(matches[1])
	}
	hours, _ := strconv.Atoi(matches[2])
	minutes, _ := strconv.Atoi(matches[3])
	seconds, _ := strconv.Atoi(matches[4])

	total := time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second

	return total, nil
}
