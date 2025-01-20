# Repository Activity Ranking Algorithm

This project implements an algorithm to rank repositories based on their activity from a CSV file containing commit data. Below is a comprehensive guide to understanding, running, and reasoning behind the implementation.

## Overview

The algorithm reads commit data from a CSV file, calculates an activity score for each repository based on the number of files changed, additions, and deletions per commit, and then ranks them accordingly. Here's a breakdown of the process:

1. **Reading CSV Data**: The CSV file is read line by line, skipping the header row to avoid parsing errors.
2. **Activity Score Calculation**: For each commit, an activity score is calculated using the formula:

```
$$
\displaystyle\AS = \text{Files} \times \text{Additions} + \text{Deletions}
$$
```
   
   Where:
   - **AS (Activity Score)**: The final score used for ranking.
   - **Files**: Number of files changed in the commit.
   - **Additions**: Number of lines added in the commit.
   - **Deletions**: Number of lines deleted in the commit.
3. **Ranking**: Repositories are sorted based on their total activity scores in descending order.
4. **Output**: The top `TopR` (currently set to 10) most active repositories are printed.

## Files

- **main.go**: Contains the main function and handles reading the CSV, processing data concurrently, and ranking repositories.
- **model.go**: Defines the `RankedRepo` struct for storing repository names and their activity scores, and sets the `TopR` constant.

## Running the Implementation

To run this Go program, follow these steps:

### Execution

1. **Compile and Run**: From within the `Activity` directory, compile and run the program:

```sh
go build main.go model.go
```

2. **Run the Program**: Execute the compiled binary with the CSV file path as an argument:

- **On Unix-like systems**:

```sh
./main "file/commits.csv"
```

- **On Windows**:

```sh
main.exe "file\commits.csv"
```

3. Alternative: Directly Running with `go run`: If you prefer not to compile separately, you can use:

```sh
go run main.go model.go "file/commits.csv"
```

### Concurrency

The implementation uses Go's goroutines for concurrent processing of CSV records to improve performance, especially in a future with a large dataset. Here's how concurrency is implemented:

- **Single Reader:** One goroutine reads the CSV file line by line.
- **Worker Pool:** Multiple worker goroutines (4 by default) process each line from a channel, calculate the score, and update a shared map with synchronization via a mutex.

## Future Implementations

While the current implementation provides a solid foundation for ranking repository activity, there are several enhancements and features that could be considered for future development:

- **Refined Scoring Formula**: 
  - **Weighted Fields**: Introduce different weights for `files`, `additions`, and `deletions` based on their significance.

- **Advanced Concurrency**:
  - **Dynamic Worker Pool**: Implement a dynamic worker pool that scales based on the system's capabilities or the size of the dataset for better performance optimization.

- **Testing Enhancements**:
  - **Integration Tests**: Add more comprehensive integration tests, including different types of CSV data.
  - **Benchmarking**: Performance benchmarking to understand the impact of concurrency and data size on execution time.

- **Command Line Interface (CLI) Improvements**:
  - **More Options**: Add command-line flags for customization like changing the number of top repositories to display or the amount of goroutines we use.
