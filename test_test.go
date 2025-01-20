package main

import (
	"encoding/csv"
	"os"
	"strings"
	"testing"
)

func TestReadCSVConcurrently(t *testing.T) {
	// Create a temporary CSV file
	tmpfile, err := os.CreateTemp("", "testcsv*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data to the temporary file
	testData := `timestamp,username,repository,files,additions,deletions
				2023-01-01,user1,repo1,1,10,5
				2023-01-02,user2,repo1,2,20,10
				2023-01-03,user3,repo2,3,30,15
				`
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Open the test file
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Test function
	repos := make(map[string]RankedRepo)
	result, err := readCSVConcurrently(reader, repos)
	if err != nil {
		t.Errorf("readCSVConcurrently failed: %v", err)
	}

	// Expected results
	expected := map[string]RankedRepo{
		"repo1": {Score: 1*10 + 5 + 2*20 + 10},
		"repo2": {Score: 3*30 + 15},
	}

	// Compare results
	for repo, data := range expected {
		if got, ok := result[repo]; !ok {
			t.Errorf("Repository %s not found in result", repo)
		} else if got != data {
			t.Errorf("For repository %s, got score %v, want %v", repo, got, data)
		}
	}
}

func TestReadCSVFileNotFound(t *testing.T) {
	// Initialize an empty map to store repositories
	repos := make(map[string]RankedRepo)
	// Test with a non-existent file
	reader := csv.NewReader(strings.NewReader(""))
	_, err := readCSVConcurrently(reader, repos)
	if err == nil {
		t.Error("Expected an error for file not found, but got none")
	}
}

func TestReadCSVEmpty(t *testing.T) {
	var repos map[string]RankedRepo
	// Test with an empty CSV
	reader := csv.NewReader(strings.NewReader("timestamp,username,repository,files,additions,deletions\n"))
	result, err := readCSVConcurrently(reader, repos)
	if err != nil {
		t.Errorf("readCSVConcurrently failed with empty CSV: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty map, but got %v", result)
	}
}

func TestReadCSVMalformed(t *testing.T) {
	var repos map[string]RankedRepo
	// Test with malformed CSV data
	reader := csv.NewReader(strings.NewReader("timestamp,username,repository,files,additions,deletions\n2023-01-01,user1,repo1,,10,5"))
	result, err := readCSVConcurrently(reader, repos)
	if err != nil {
		t.Errorf("readCSVConcurrently failed with malformed CSV: %v", err)
	}
	// Since we skip malformed lines, the result should be empty
	if len(result) != 0 {
		t.Errorf("Expected empty map due to malformed data, but got %v", result)
	}
}

func TestCalculateActivityScore(t *testing.T) {
	testCases := []struct {
		files     int
		additions int
		deletions int
		score     int
	}{
		{1, 10, 5, 1*10 + 5},
		{2, 20, 10, 2*20 + 10},
		{3, 30, 15, 3*30 + 15},
	}

	for _, tc := range testCases {
		got := calculateActivityScore(tc.files, tc.additions, tc.deletions)
		if got != tc.score {
			t.Errorf("For files %d, additions %d, deletions %d, got score %d, want %d", tc.files, tc.additions, tc.deletions, got, tc.score)
		}
	}
}
