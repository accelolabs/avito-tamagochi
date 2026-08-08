package game

import "time"

const MoscowLocation = "Europe/Moscow"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

func MoscowDate(now time.Time) time.Time {
	location, err := time.LoadLocation(MoscowLocation)
	if err != nil {
		location = time.FixedZone("MSK", 3*60*60)
	}
	local := now.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}
