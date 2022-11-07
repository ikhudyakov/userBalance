package utility

import "time"

func EndOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, 0).Add(-time.Nanosecond)
}
