package services

import (
	"time"

	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
)

type StreakInfo struct {
	Current int `json:"current"`
	Longest int `json:"longest"`
}

type WeeklyProgress struct {
	Day     string `json:"day"`
	Date    string `json:"date"`
	Minutes int    `json:"minutes"`
}

type YearlyProgress struct {
	Month   string  `json:"month"`
	Hours   float64 `json:"hours"`
	Minutes int     `json:"minutes"`
}

type DashboardData struct {
	User            models.User      `json:"user"`
	Streaks         StreakInfo       `json:"streaks"`
	WeeklyProgress  []WeeklyProgress `json:"weekly_progress"`
	YearlyProgress  []YearlyProgress `json:"yearly_progress"`
	RecentSessions  []models.Session `json:"recent_sessions"`
}

// CalculateStreaks calculates current and longest streak for a user using efficient SQL queries
func CalculateStreaks(userID string) (StreakInfo, error) {
	sessionDates, err := getSessionDates(userID)
	if err != nil {
		return StreakInfo{}, err
	}

	longestStreak := calculateLongestStreak(sessionDates)
	currentStreak := calculateCurrentStreak(sessionDates)

	return StreakInfo{
		Current: currentStreak,
		Longest: longestStreak,
	}, nil
}

// calculateLongestStreak calculates the longest streak from session dates
func calculateLongestStreak(sessionDates []string) int {
	if len(sessionDates) == 0 {
		return 0
	}

	// Need ascending order for longest streak calculation
	// Reverse the descending order array
	ascending := make([]string, len(sessionDates))
	for i, date := range sessionDates {
		ascending[len(sessionDates)-1-i] = date
	}

	// Calculate longest streak from session dates
	longestStreak := 1
	currentStreak := 1
	
	for i := 1; i < len(ascending); i++ {
		prevDate, _ := time.Parse("2006-01-02", ascending[i-1])
		currDate, _ := time.Parse("2006-01-02", ascending[i])
		
		// Check if dates are consecutive
		if currDate.Sub(prevDate).Hours() == 24 {
			currentStreak++
		} else {
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
			currentStreak = 1
		}
	}
	
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}
	
	return longestStreak
}

// calculateCurrentStreak calculates current streak from session dates (already in DESC order)
func calculateCurrentStreak(sessionDates []string) int {
	if len(sessionDates) == 0 {
		return 0
	}

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	
	// Check if we should start counting from today or yesterday
	startDate := ""
	if sessionDates[0] == today {
		startDate = today
	} else if sessionDates[0] == yesterday {
		startDate = yesterday
	} else {
		// No recent sessions, no current streak
		return 0
	}

	// Count consecutive days backwards from start date
	currentStreak := 0
	expectedDate, _ := time.Parse("2006-01-02", startDate)
	
	for _, dateStr := range sessionDates {
		sessionDate, _ := time.Parse("2006-01-02", dateStr)
		
		// Check if this date matches our expected consecutive date
		if sessionDate.Format("2006-01-02") == expectedDate.Format("2006-01-02") {
			currentStreak++
			expectedDate = expectedDate.AddDate(0, 0, -1)
		} else {
			// Gap found, streak ends
			break
		}
	}
	
	return currentStreak
}

// hasSessionOnDate checks if user has any session on a specific date
func hasSessionOnDate(userID string, date time.Time) (bool, error) {
	var count int64
	dateStr := date.Format("2006-01-02")

	err := database.DB.Model(&models.Session{}).
		Where("user_id = ? AND DATE(created_at) = ? AND deleted_at IS NULL", userID, dateStr).
		Count(&count).Error

	return count > 0, err
}

// GetWeeklyProgress gets the last 7 days of meditation progress
func GetWeeklyProgress(userID string) ([]WeeklyProgress, error) {
	var progress []WeeklyProgress

	// Get last 7 days
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		dayName := date.Format("Mon")

		var totalSeconds int
		err := database.DB.Model(&models.Session{}).
			Where("user_id = ? AND DATE(created_at) = ? AND deleted_at IS NULL", userID, dateStr).
			Select("COALESCE(SUM(duration_seconds), 0)").
			Scan(&totalSeconds).Error
		if err != nil {
			return nil, err
		}
		
		totalMinutes := totalSeconds / 60

		progress = append(progress, WeeklyProgress{
			Day:     dayName,
			Date:    dateStr,
			Minutes: totalMinutes,
		})
	}

	return progress, nil
}

// GetYearlyProgress gets monthly meditation progress for the specified year
func GetYearlyProgress(userID string, year int) ([]YearlyProgress, error) {
	var progress []YearlyProgress

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for i, month := range months {
		monthStart := time.Date(year, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var totalSeconds int
		err := database.DB.Model(&models.Session{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL",
				userID, monthStart, monthEnd).
			Select("COALESCE(SUM(duration_seconds), 0)").
			Scan(&totalSeconds).Error
		if err != nil {
			return nil, err
		}

		minutes := totalSeconds / 60
		hours := float64(totalSeconds) / 3600.0

		progress = append(progress, YearlyProgress{
			Month:   month,
			Hours:   hours,
			Minutes: minutes,
		})
	}

	return progress, nil
}

// GetRecentSessions gets recent sessions for a user with configurable limit
func GetRecentSessions(userID string, limit int) ([]models.Session, error) {
	var sessions []models.Session

	// Set default limit if not provided or invalid
	if limit <= 0 || limit > 100 {
		limit = 5
	}

	err := database.DB.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&sessions).Error

	return sessions, err
}

// GetDashboardData aggregates all dashboard data for a user with configurable parameters
func GetDashboardData(user *models.User, year int, sessionLimit int) (*DashboardData, error) {
	streaks, err := CalculateStreaks(user.ID)
	if err != nil {
		return nil, err
	}

	weeklyProgress, err := GetWeeklyProgress(user.ID)
	if err != nil {
		return nil, err
	}

	// Default to current year if not provided
	if year <= 0 {
		year = time.Now().Year()
	}

	yearlyProgress, err := GetYearlyProgress(user.ID, year)
	if err != nil {
		return nil, err
	}

	recentSessions, err := GetRecentSessions(user.ID, sessionLimit)
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		User:            *user,
		Streaks:         streaks,
		WeeklyProgress:  weeklyProgress,
		YearlyProgress:  yearlyProgress,
		RecentSessions:  recentSessions,
	}, nil
}

// getSessionDates retrieves distinct session dates for a user in descending order
func getSessionDates(userID string) ([]string, error) {
	var sessionDates []string
	err := database.DB.Model(&models.Session{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("DISTINCT DATE(created_at) as session_date").
		Order("session_date DESC").
		Pluck("session_date", &sessionDates).Error
	
	return sessionDates, err
}