// nolint:revive
package util

import (
	"fmt"
	"time"

	"github.com/go-openapi/swag"
)

const (
	// DateFormat defines the standard date format (YYYY-MM-DD).
	DateFormat       = "2006-01-02"
	monthsPerQuarter = 3
	daysPerWeek      = 7
)

// TimeFromString parses a time string assuming RFC3339 format.
//
// Parameters:
//   - timeString: The string to parse.
//
// Returns:
//   - time.Time: The parsed time.
//   - error: An error if parsing fails.
func TimeFromString(timeString string) (time.Time, error) {
	result, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time string: %w", err)
	}

	return result, nil
}

// DateFromString parses a date string assuming YYYY-MM-DD format.
//
// Parameters:
//   - dateString: The string to parse.
//
// Returns:
//   - time.Time: The parsed date.
//   - error: An error if parsing fails.
func DateFromString(dateString string) (time.Time, error) {
	result, err := time.Parse(DateFormat, dateString)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date string: %w", err)
	}

	return result, nil
}

// EndOfMonth returns the last nanosecond of the month for the given time.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The end of the month.
func EndOfMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, -1, d.Location())
}

// EndOfPreviousMonth returns the last nanosecond of the previous month.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The end of the previous month.
func EndOfPreviousMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, -1, d.Location())
}

// EndOfDay returns the last nanosecond of the day for the given time.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The end of the day.
func EndOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day()+1, 0, 0, 0, -1, d.Location())
}

// StartOfDay returns the first nanosecond (midnight) of the day for the given time.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The start of the day.
func StartOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// StartOfMonth returns the first nanosecond of the month for the given time.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The start of the month.
func StartOfMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}

// StartOfQuarter returns the first nanosecond of the quarter for the given time.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The start of the quarter.
func StartOfQuarter(d time.Time) time.Time {
	quarter := (int(d.Month()) - 1) / monthsPerQuarter
	m := quarter*monthsPerQuarter + 1
	return time.Date(d.Year(), time.Month(m), 1, 0, 0, 0, 0, d.Location())
}

// StartOfWeek returns the Monday (assuming week starts on Monday) of the week for the given date.
//
// Parameters:
//   - date: The reference time.
//
// Returns:
//   - time.Time: The start of the week.
func StartOfWeek(date time.Time) time.Time {
	dayOffset := int(date.Weekday()) - 1

	// go time is starting weeks at sunday
	if dayOffset < 0 {
		dayOffset = 6
	}

	return time.Date(date.Year(), date.Month(), date.Day()-dayOffset, 0, 0, 0, 0, date.Location())
}

// Date constructs a time.Time for the given date at midnight.
//
// Parameters:
//   - year: The year.
//   - month: The month.
//   - day: The day.
//   - loc: The time location.
//
// Returns:
//   - time.Time: The constructed date.
func Date(year int, month int, day int, loc *time.Location) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
}

// AddWeeks adds a number of weeks to the given time.
//
// Parameters:
//   - d: The reference time.
//   - weeks: The number of weeks to add.
//
// Returns:
//   - time.Time: The resulting time.
func AddWeeks(d time.Time, weeks int) time.Time {
	return d.AddDate(0, 0, daysPerWeek*weeks)
}

// AddMonths adds a number of months to the given time.
//
// Parameters:
//   - d: The reference time.
//   - months: The number of months to add.
//
// Returns:
//   - time.Time: The resulting time.
func AddMonths(d time.Time, months int) time.Time {
	return d.AddDate(0, months, 0)
}

// DayBefore returns the last nanosecond of the previous day.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The end of the previous day.
func DayBefore(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, -1, d.Location())
}

// TruncateTime removes the time component, returning midnight of the same day.
//
// Parameters:
//   - d: The reference time.
//
// Returns:
//   - time.Time: The truncated time.
func TruncateTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// MaxTime returns the latest time from a list of times.
//
// Parameters:
//   - times: A variadic list of time.Time.
//
// Returns:
//   - time.Time: The latest time in the list.
func MaxTime(times ...time.Time) time.Time {
	var latestTime time.Time
	for _, t := range times {
		if t.After(latestTime) {
			latestTime = t
		}
	}

	return latestTime
}

// NonZeroTimeOrNil returns a pointer to the passed time if it is not a zero time.
// Passing a zero/uninitialized time returns nil.
//
// Parameters:
//   - t: The time to check.
//
// Returns:
//   - *time.Time: A pointer to the time or nil.
func NonZeroTimeOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}

	return swag.Time(t)
}
