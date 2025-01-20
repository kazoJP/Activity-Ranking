package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

func main() {
	// check if a file path was provided
	if len(os.Args) != 2 {
		log.Fatal("Please provide a CSV file path as an argument. Usage: go run main.go <file.csv>")
	}

	// the CSV file path is the second argument (index 1)
	filename := os.Args[1]

	cleanpath, err := readFile(filename) // check the file path
	if err != nil {
		log.Printf("%v", err) // log the error and exit
		return
	}

	file, err := os.Open(cleanpath) // open the file
	if err != nil {
		log.Printf("Error opening file %s: %v", cleanpath, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file) // create a new reader
	reader.FieldsPerRecord = -1   // set to -1 to allow variable number of fields

	repos := make(map[string]RankedRepo)                   // repositiory name -> RankedRepo struct
	repActivity, err := readCSVConcurrently(reader, repos) // read the CSV file into the map
	if err != nil {
		log.Printf("%v", err) // log the error and exit
		return
	}

	ranked, err := rankRepositories(repActivity) // rank the repositories
	if err != nil {
		log.Printf("Error ranking repositories: %v", err)
		return
	}

	// Print the top 'TopR' ranked repositories
	for i, repo := range ranked {
		if i >= TopR {
			break
		}
		fmt.Printf("%d. %s - Score: %d\n", i+1, repo.Name, repo.Score)
	}
}

// cleans and checks the file path, ensuring it exists and is a .csv file
func readFile(filePath string) (string, error) {
	// Clean the file path to prevent directory traversal attacks
	cleanPath := filepath.Clean(filePath)

	// Check if the file exists and is accessible
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", cleanPath)
	} else if err != nil {
		return "", fmt.Errorf("error checking file: %v", cleanPath)
	}

	// Ensure it's a .csv file
	if filepath.Ext(cleanPath) != ".csv" {
		return "", fmt.Errorf("file must have .csv extension: %s", cleanPath)
	}
	return cleanPath, nil
}

// readCSV reads the CSV file and returns a map of repositories and their scores
func readCSVConcurrently(reader *csv.Reader, repos map[string]RankedRepo) (map[string]RankedRepo, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	recordsChan := make(chan []string, 100)

	numWorkers := 4
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go worker(&wg, &mu, recordsChan, repos)
	}

	_, err := reader.Read() // Skip header
	if err != nil {
		close(recordsChan)
		return nil, fmt.Errorf("error reading CSV header: %w", err)
	}

	// Read the CSV file line by line
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			close(recordsChan)
			return nil, fmt.Errorf("error reading CSV: %w", err)
		}

		// Check if the record has the required fields
		if len(record) < 6 || record[2] == "" || record[3] == "" || record[4] == "" || record[5] == "" {
			log.Printf("Skipping line due to missing required fields: %v", record)
			continue
		}

		recordsChan <- record
	}

	// Close the channel and wait for all workers to finish
	close(recordsChan)
	wg.Wait()

	return repos, nil
}

// worker processes records from the recordsChan and updates the repositories map
func worker(wg *sync.WaitGroup, mu *sync.Mutex, recordsChan <-chan []string, repos map[string]RankedRepo) {
	defer wg.Done()
	for record := range recordsChan {
		repoName := record[2]

		processField := func(fieldName string, value string) (int, bool) {
			num, err := strconv.Atoi(value)
			if err != nil || num < 0 {
				log.Printf("Skipping line for repo %s due to invalid %s value: %v, record: %v", repoName, fieldName, err, record)
				return 0, false
			}
			return num, true
		}
		files, ok := processField("files", record[3])
		if !ok {
			continue
		}
		additions, ok := processField("additions", record[4])
		if !ok {
			continue
		}
		deletions, ok := processField("deletions", record[5])
		if !ok {
			continue
		}

		score := calculateActivityScore(files, additions, deletions)

		mu.Lock()
		if existing, found := repos[repoName]; found {
			existing.Score += score
			repos[repoName] = existing
		} else {
			repos[repoName] = RankedRepo{Score: score}
		}
		mu.Unlock()
	}
}

// rankRepositories calculates an activity score for each repository and returns them sorted by score
func rankRepositories(repos map[string]RankedRepo) ([]struct {
	Name  string
	Score int
}, error) {
	// convert the map to a slice of structs for sorting
	var rankedRepos []struct {
		Name  string
		Score int
	}
	for name, repo := range repos {
		rankedRepos = append(rankedRepos, struct {
			Name  string
			Score int
		}{Name: name, Score: repo.Score})
	}

	// sort the slice by score in descending order
	sort.Slice(rankedRepos, func(i, j int) bool {
		return rankedRepos[i].Score > rankedRepos[j].Score
	})

	return rankedRepos, nil
}

// calculate ActivityScore(AS) computes an activity score for a repository
func calculateActivityScore(files int, additions int, deletions int) int {
	return files*additions + deletions
}
