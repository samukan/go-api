package utils

import "time"

// AgeFromBirthdate computes age in full years from a YYYY-MM-DD date string.
// Returns -1 if parsing fails.
func AgeFromBirthdate(s string) int {
    if s == "" {
        return -1
    }
    t, err := time.Parse("2006-01-02", s)
    if err != nil {
        return -1
    }
    now := time.Now().UTC()
    years := now.Year() - t.Year()
    // If birthday hasn't happened yet this year, subtract 1
    bday := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
    if now.Before(bday) {
        years--
    }
    if years < 0 {
        return -1
    }
    return years
}

// AgeFromTime computes age in full years from a time.Time birthdate.
func AgeFromTime(t time.Time) int {
    if t.IsZero() {
        return -1
    }
    now := time.Now().UTC()
    years := now.Year() - t.Year()
    bday := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
    if now.Before(bday) {
        years--
    }
    if years < 0 {
        return -1
    }
    return years
}
