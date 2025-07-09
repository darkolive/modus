package helpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestUser represents a test user for unit and integration tests
type TestUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	CentreID string `json:"centreId"`
}

// TestCourse represents a test course for unit and integration tests
type TestCourse struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	CentreID    string       `json:"centreId"`
	Status      string       `json:"status"`
	Modules     []TestModule `json:"modules"`
	CreatedAt   string       `json:"createdAt"`
	CreatedBy   string       `json:"createdBy"`
}

// TestModule represents a test module for unit and integration tests
type TestModule struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Order  int    `json:"order"`
	Status string `json:"status"`
}

// LoadTestUsers loads test user data from the mock JSON file
func LoadTestUsers(t *testing.T) []TestUser {
	data, err := loadTestData("mock_users.json")
	if err != nil {
		t.Fatalf("Failed to load test users: %v", err)
	}

	var result struct {
		Users []TestUser `json:"users"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse test users: %v", err)
	}

	return result.Users
}

// LoadTestCourses loads test course data from the mock JSON file
func LoadTestCourses(t *testing.T) []TestCourse {
	data, err := loadTestData("mock_courses.json")
	if err != nil {
		t.Fatalf("Failed to load test courses: %v", err)
	}

	var result struct {
		Courses []TestCourse `json:"courses"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse test courses: %v", err)
	}

	return result.Courses
}

// loadTestData loads test data from the testdata directory
func loadTestData(filename string) ([]byte, error) {
	// Calculate path to testdata directory relative to this file
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "testdata")
	
	return os.ReadFile(filepath.Join(testdataDir, filename))
}
