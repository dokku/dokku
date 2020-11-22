package common

import (
	"fmt"
	"runtime"
	"sync"
)

// ParallelCommand is a type that declares functions
// the ps plugins can execute in parallel
type ParallelCommand func(string) error

// ParallelCommandRun contains all the arguments
// necessary for a parallel command run
type ParallelCommandRun struct {
	Name string
}

// ParallelCommandResult is the result of a parallel
// command run
type ParallelCommandResult struct {
	Name  string
	Error error
}

// RunCommandAgainstAllApps runs a given ParallelCommand against all apps
func RunCommandAgainstAllApps(command ParallelCommand, commandName string, parallelCount int) error {
	runInSerial := false

	if parallelCount < -1 {
		return fmt.Errorf("Invalid value %d for --parallel flag", parallelCount)
	}

	if parallelCount == -1 {
		cpuCount := runtime.NumCPU()
		LogWarn(fmt.Sprintf("Setting --parallel=%d value to CPU count of %d", parallelCount, cpuCount))
		parallelCount = cpuCount
	}

	if parallelCount == 0 || parallelCount == 1 {
		LogWarn(fmt.Sprintf("Running %s in serial mode", commandName))
		runInSerial = true
	}

	if runInSerial {
		return runCommandAgainstAllAppsSerially(command, commandName)
	}

	return runCommandAgainstAllAppsInParallel(command, commandName, parallelCount)
}

func runCommandAgainstAllAppsInParallel(command ParallelCommand, commandName string, parallelCount int) error {
	apps, err := DokkuApps()
	if err != nil {
		LogWarn(err.Error())
		return nil
	}

	jobs := make(chan string, parallelCount)
	results := make(chan ParallelCommandResult, len(apps))

	go allocateJobs(apps, jobs)
	done := make(chan error)
	go aggregateResults(results, done)
	createParallelWorkerPool(jobs, results, command, parallelCount)
	err = <-done

	return err
}

func runCommandAgainstAllAppsSerially(command ParallelCommand, commandName string) error {
	apps, err := DokkuApps()
	if err != nil {
		LogWarn(err.Error())
		return nil
	}

	errorCount := 0
	for _, appName := range apps {
		LogInfo1(fmt.Sprintf("Running %s against app %s", commandName, appName))
		if err = command(appName); err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%s command returned %d errors", commandName, errorCount)
	}

	return nil
}

func allocateJobs(input []string, jobs chan string) {
	for _, job := range input {
		jobs <- job
	}
	close(jobs)
}

func aggregateResults(results chan ParallelCommandResult, done chan error) {
	var parallelError error
	errorCount := 0
	for result := range results {
		if result.Error != nil {
			LogWarn(fmt.Sprintf("Error running command against %s", result.Name))
			errorCount++
		}
	}
	if errorCount > 0 {
		parallelError = fmt.Errorf("Encountered %d errors during parallel run", errorCount)
	}
	done <- parallelError
}

func createParallelWorker(jobs chan string, results chan ParallelCommandResult, command ParallelCommand, wg *sync.WaitGroup, workerID int) {
	for job := range jobs {
		LogInfo1(fmt.Sprintf("Running command against %s", job))
		output := ParallelCommandResult{
			Name:  job,
			Error: command(job),
		}
		results <- output
	}
	wg.Done()
}

func createParallelWorkerPool(jobs chan string, results chan ParallelCommandResult, command ParallelCommand, numberOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go createParallelWorker(jobs, results, command, &wg, i)
	}
	wg.Wait()
	close(results)
}
