package pretty

import (
	"strings"
	"testing"
	"time"
)

func TestTimeFormatter_Format(t *testing.T) {
	// Fixed reference time for consistent testing
	now := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	tf := NewTimeFormatter().WithNow(now)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "<zero>",
		},
		{
			name:     "just now - under 10 seconds",
			input:    now.Add(-5 * time.Second),
			expected: "just now",
		},
		{
			name:     "seconds ago",
			input:    now.Add(-30 * time.Second),
			expected: "30 seconds ago",
		},
		{
			name:     "1 second ago",
			input:    now.Add(-1 * time.Second),
			expected: "just now", // Under 10 seconds with friendly phrases
		},
		{
			name:     "1 minute ago",
			input:    now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			input:    now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			input:    now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			input:    now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "yesterday",
			input:    now.Add(-24 * time.Hour),
			expected: "yesterday",
		},
		{
			name:     "2 days ago",
			input:    now.Add(-48 * time.Hour),
			expected: "2 days ago",
		},
		{
			name:     "last week",
			input:    now.Add(-7 * 24 * time.Hour),
			expected: "last week",
		},
		{
			name:     "2 weeks ago",
			input:    now.Add(-14 * 24 * time.Hour),
			expected: "2 weeks ago",
		},
		{
			name:     "last month",
			input:    now.Add(-30 * 24 * time.Hour),
			expected: "last month",
		},
		{
			name:     "2 months ago",
			input:    now.Add(-60 * 24 * time.Hour),
			expected: "2 months ago",
		},
		{
			name:     "last year",
			input:    now.Add(-365 * 24 * time.Hour),
			expected: "last year",
		},
		{
			name:     "2 years ago",
			input:    now.Add(-2 * 365 * 24 * time.Hour),
			expected: "2 years ago",
		},
		// Future times
		{
			name:     "in 5 seconds",
			input:    now.Add(5 * time.Second),
			expected: "just now", // Under 10 seconds treated as "just now"
		},
		{
			name:     "in 30 seconds",
			input:    now.Add(30 * time.Second),
			expected: "in 30 seconds",
		},
		{
			name:     "in 1 minute",
			input:    now.Add(1 * time.Minute),
			expected: "in 1 minute",
		},
		{
			name:     "in 5 minutes",
			input:    now.Add(5 * time.Minute),
			expected: "in 5 minutes",
		},
		{
			name:     "in 1 hour",
			input:    now.Add(1 * time.Hour),
			expected: "in 1 hour",
		},
		{
			name:     "tomorrow",
			input:    now.Add(24 * time.Hour),
			expected: "tomorrow",
		},
		{
			name:     "in 2 days",
			input:    now.Add(48 * time.Hour),
			expected: "in 2 days",
		},
		{
			name:     "next week",
			input:    now.Add(7 * 24 * time.Hour),
			expected: "next week",
		},
		{
			name:     "next month",
			input:    now.Add(30 * 24 * time.Hour),
			expected: "next month",
		},
		{
			name:     "next year",
			input:    now.Add(365 * 24 * time.Hour),
			expected: "next year",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTimeFormatter_WithoutFriendlyPhrases(t *testing.T) {
	now := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	tf := NewTimeFormatter().WithNow(now).WithFriendlyPhrases(false)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "5 seconds ago - no friendly phrase",
			input:    now.Add(-5 * time.Second),
			expected: "5 seconds ago",
		},
		{
			name:     "1 day ago - no friendly phrase",
			input:    now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "1 week ago - no friendly phrase",
			input:    now.Add(-7 * 24 * time.Hour),
			expected: "1 week ago",
		},
		{
			name:     "1 month ago - no friendly phrase",
			input:    now.Add(-30 * 24 * time.Hour),
			expected: "1 month ago",
		},
		{
			name:     "1 year ago - no friendly phrase",
			input:    now.Add(-365 * 24 * time.Hour),
			expected: "1 year ago",
		},
		// Future times
		{
			name:     "in 1 day - no friendly phrase",
			input:    now.Add(24 * time.Hour),
			expected: "in 1 day",
		},
		{
			name:     "in 1 week - no friendly phrase",
			input:    now.Add(7 * 24 * time.Hour),
			expected: "in 1 week",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTimeFormatter_CustomFutureFormat(t *testing.T) {
	now := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	tf := NewTimeFormatter().WithNow(now).WithFutureFormat("%s from now").WithFriendlyPhrases(false)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "30 seconds from now",
			input:    now.Add(30 * time.Second),
			expected: "30 seconds from now",
		},
		{
			name:     "5 minutes from now",
			input:    now.Add(5 * time.Minute),
			expected: "5 minutes from now",
		},
		{
			name:     "2 hours from now",
			input:    now.Add(2 * time.Hour),
			expected: "2 hours from now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTimeFormatter_CustomThresholds(t *testing.T) {
	now := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	// Custom formatter with very low second threshold (10 seconds)
	tf := NewTimeFormatter().
		WithNow(now).
		WithSecondThreshold(10 * time.Second).
		WithFriendlyPhrases(false)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "5 seconds ago - under threshold",
			input:    now.Add(-5 * time.Second),
			expected: "5 seconds ago",
		},
		{
			name:     "15 seconds ago - over threshold, should be minutes",
			input:    now.Add(-15 * time.Second),
			expected: "0 minutes ago", // Rounds down to 0 minutes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTimeFormatter_ChainedMethods(t *testing.T) {
	now := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	// Test method chaining
	tf := NewTimeFormatter().
		WithNow(now).
		WithFriendlyPhrases(false).
		WithFutureFormat("%s from now").
		WithSecondThreshold(2 * time.Minute).
		WithMinuteThreshold(2 * time.Hour)

	// Verify all settings took effect
	if tf.Now != now {
		t.Errorf("WithNow() not applied correctly")
	}
	if tf.FriendlyPhrases != false {
		t.Errorf("WithFriendlyPhrases() not applied correctly")
	}
	if tf.FutureFormat != "%s from now" {
		t.Errorf("WithFutureFormat() not applied correctly")
	}
	if tf.SecondThreshold != 2*time.Minute {
		t.Errorf("WithSecondThreshold() not applied correctly")
	}
	if tf.MinuteThreshold != 2*time.Hour {
		t.Errorf("WithMinuteThreshold() not applied correctly")
	}
}

func TestTime_GlobalFunction(t *testing.T) {
	// Test the global Time function (uses current time as reference)
	past := time.Now().Add(-5 * time.Minute)
	result := Time(past)

	// Should contain "minutes ago" for a time 5 minutes in the past
	if !strings.Contains(result, "minutes ago") && !strings.Contains(result, "minute ago") {
		t.Errorf("Time() = %q, expected to contain 'minute(s) ago'", result)
	}

	// Test zero time
	zeroResult := Time(time.Time{})
	if zeroResult != "<zero>" {
		t.Errorf("Time(zero) = %q, want '<zero>'", zeroResult)
	}
}

func TestTimeFormatterDefaults(t *testing.T) {
	tf := NewTimeFormatter()

	// Test default values
	if tf.SecondThreshold != 1*time.Minute {
		t.Errorf("Default SecondThreshold = %v, want %v", tf.SecondThreshold, 1*time.Minute)
	}
	if tf.MinuteThreshold != 1*time.Hour {
		t.Errorf("Default MinuteThreshold = %v, want %v", tf.MinuteThreshold, 1*time.Hour)
	}
	if tf.HourThreshold != 24*time.Hour {
		t.Errorf("Default HourThreshold = %v, want %v", tf.HourThreshold, 24*time.Hour)
	}
	if tf.DayThreshold != 7*24*time.Hour {
		t.Errorf("Default DayThreshold = %v, want %v", tf.DayThreshold, 7*24*time.Hour)
	}
	if tf.WeekThreshold != 30*24*time.Hour {
		t.Errorf("Default WeekThreshold = %v, want %v", tf.WeekThreshold, 30*24*time.Hour)
	}
	if tf.MonthThreshold != 365*24*time.Hour {
		t.Errorf("Default MonthThreshold = %v, want %v", tf.MonthThreshold, 365*24*time.Hour)
	}
	if tf.FriendlyPhrases != true {
		t.Errorf("Default UseFriendlyPhrases = %v, want true", tf.FriendlyPhrases)
	}
	if tf.FutureFormat != "in %s" {
		t.Errorf("Default FutureFormat = %q, want 'in %%s'", tf.FutureFormat)
	}
}

func TestPrinterTimeIntegration(t *testing.T) {
	// Test that the Printer integrates with time formatting
	printer := New().WithColorMode(ColorNever)

	// Test with a struct containing a time field
	data := struct {
		Name      string
		CreatedAt time.Time
	}{
		Name:      "Test",
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}

	result := printer.Print(data)

	// Should contain the time formatted as relative time
	if !strings.Contains(result, "minutes ago") && !strings.Contains(result, "minute ago") {
		t.Errorf("Expected result to contain formatted time, got: %s", result)
	}

	// Test with zero time
	zeroData := struct {
		Time time.Time
	}{
		Time: time.Time{},
	}

	zeroResult := printer.Print(zeroData)
	if !strings.Contains(zeroResult, "<zero>") {
		t.Errorf("Expected result to contain '<zero>' for zero time, got: %s", zeroResult)
	}
}
