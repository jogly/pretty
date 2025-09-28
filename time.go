package pretty

import (
	"fmt"
	"time"
)

// TimeFormatter configures and performs human-friendly relative time formatting
type TimeFormatter struct {
	// Reference time for calculating relative time (defaults to time.Now())
	Now time.Time

	SecondThreshold time.Duration // Show "X seconds ago" below this (default: 1 minute)
	MinuteThreshold time.Duration // Show "X minutes ago" below this (default: 1 hour)
	HourThreshold   time.Duration // Show "X hours ago" below this (default: 1 day)
	DayThreshold    time.Duration // Show "X days ago" below this (default: 1 week)
	WeekThreshold   time.Duration // Show "X weeks ago" below this (default: 1 month)
	MonthThreshold  time.Duration // Show "X months ago" below this (default: 1 year)

	// Use friendly phrases like "just now", "last week", "next month"
	FriendlyPhrases bool

	// Show future times as "in X time" vs "X from now".
	// "in %s" (default), or customize like "%s from now"
	FutureFormat string

	ZeroString string // String to show for zero time (default: "<zero>")
}

// NewTimeFormatter creates a new TimeFormatter with sensible defaults
func NewTimeFormatter() *TimeFormatter {
	return &TimeFormatter{
		Now:             time.Now(),
		SecondThreshold: 1 * time.Minute,
		MinuteThreshold: 1 * time.Hour,
		HourThreshold:   24 * time.Hour,
		DayThreshold:    7 * 24 * time.Hour,
		WeekThreshold:   30 * 24 * time.Hour,
		MonthThreshold:  365 * 24 * time.Hour,
		FriendlyPhrases: true,
		FutureFormat:    "in %s",
		ZeroString:      "<zero>",
	}
}

// WithNow sets a custom reference time for relative calculations
func (tf *TimeFormatter) WithNow(now time.Time) *TimeFormatter {
	newTF := *tf
	newTF.Now = now
	return &newTF
}

// WithSecondThreshold sets when to stop showing seconds and switch to minutes
func (tf *TimeFormatter) WithSecondThreshold(d time.Duration) *TimeFormatter {
	newTF := *tf
	newTF.SecondThreshold = d
	return &newTF
}

// WithMinuteThreshold sets when to stop showing minutes and switch to hours
func (tf *TimeFormatter) WithMinuteThreshold(d time.Duration) *TimeFormatter {
	newTF := *tf
	newTF.MinuteThreshold = d
	return &newTF
}

// WithHourThreshold sets when to stop showing hours and switch to days
func (tf *TimeFormatter) WithHourThreshold(d time.Duration) *TimeFormatter {
	newTF := *tf
	newTF.HourThreshold = d
	return &newTF
}

// WithDayThreshold sets when to stop showing days and switch to weeks
func (tf *TimeFormatter) WithDayThreshold(d time.Duration) *TimeFormatter {
	newTF := *tf
	newTF.DayThreshold = d
	return &newTF
}

// WithWeekThreshold sets when to stop showing weeks and switch to months
func (tf *TimeFormatter) WithWeekThreshold(d time.Duration) *TimeFormatter {
	newTF := *tf
	newTF.WeekThreshold = d
	return &newTF
}

// WithMonthThreshold sets when to stop showing months and switch to years
func (tf *TimeFormatter) WithMonthThreshold(d time.Duration) *TimeFormatter {
	newTF := *tf
	newTF.MonthThreshold = d
	return &newTF
}

// WithFriendlyPhrases enables/disables friendly phrases like "just now", "last week"
func (tf *TimeFormatter) WithFriendlyPhrases(enabled bool) *TimeFormatter {
	newTF := *tf
	newTF.FriendlyPhrases = enabled
	return &newTF
}

// WithFutureFormat sets how future times are formatted ("in %s" vs "%s from now")
func (tf *TimeFormatter) WithFutureFormat(format string) *TimeFormatter {
	newTF := *tf
	newTF.FutureFormat = format
	return &newTF
}

// Format formats a time.Time value into a human-friendly relative string
func (tf *TimeFormatter) Format(t time.Time) string {
	if t.IsZero() {
		return tf.ZeroString
	}

	now := tf.Now
	if now.IsZero() {
		now = time.Now()
	}

	diff := now.Sub(t)
	absDiff := diff.Abs()

	var result string
	isPast := diff > 0

	switch {
	case absDiff < tf.SecondThreshold:
		seconds := int(absDiff.Seconds())
		if tf.FriendlyPhrases && seconds < 10 {
			result = "just now"
			isPast = true // "just now" is always considered past
		} else if seconds == 1 {
			result = "1 second"
		} else {
			result = fmt.Sprintf("%d seconds", seconds)
		}

	case absDiff < tf.MinuteThreshold:
		minutes := int(absDiff.Minutes())
		if minutes == 1 {
			result = "1 minute"
		} else {
			result = fmt.Sprintf("%d minutes", minutes)
		}

	case absDiff < tf.HourThreshold:
		hours := int(absDiff.Hours())
		if hours == 1 {
			result = "1 hour"
		} else {
			result = fmt.Sprintf("%d hours", hours)
		}

	case absDiff < tf.DayThreshold:
		days := int(absDiff.Hours() / 24)
		if tf.FriendlyPhrases && days == 1 {
			if isPast {
				result = "yesterday"
			} else {
				result = "tomorrow"
			}
			isPast = true // handled by the phrase itself
		} else if days == 1 {
			result = "1 day"
		} else {
			result = fmt.Sprintf("%d days", days)
		}

	case absDiff < tf.WeekThreshold:
		weeks := int(absDiff.Hours() / (24 * 7))
		if tf.FriendlyPhrases && weeks == 1 {
			if isPast {
				result = "last week"
			} else {
				result = "next week"
			}
			isPast = true // handled by the phrase itself
		} else if weeks == 1 {
			result = "1 week"
		} else {
			result = fmt.Sprintf("%d weeks", weeks)
		}

	case absDiff < tf.MonthThreshold:
		months := int(absDiff.Hours() / (24 * 30)) // Approximate
		if tf.FriendlyPhrases && months == 1 {
			if isPast {
				result = "last month"
			} else {
				result = "next month"
			}
			isPast = true // handled by the phrase itself
		} else if months == 1 {
			result = "1 month"
		} else {
			result = fmt.Sprintf("%d months", months)
		}

	default:
		years := int(absDiff.Hours() / (24 * 365)) // Approximate
		if tf.FriendlyPhrases && years == 1 {
			if isPast {
				result = "last year"
			} else {
				result = "next year"
			}
			isPast = true // handled by the phrase itself
		} else if years == 1 {
			result = "1 year"
		} else {
			result = fmt.Sprintf("%d years", years)
		}
	}

	// Handle friendly phrases that already contain directionality
	if tf.FriendlyPhrases && (result == "just now" || result == "yesterday" || result == "tomorrow" ||
		result == "last week" || result == "next week" || result == "last month" || result == "next month" ||
		result == "last year" || result == "next year") {
		return result
	}

	// Add direction for regular phrases
	if isPast {
		return result + " ago"
	} else {
		return fmt.Sprintf(tf.FutureFormat, result)
	}
}

// Time formats a time.Time value into a human-friendly relative string using default settings
func Time(t time.Time) string {
	return NewTimeFormatter().Format(t)
}
