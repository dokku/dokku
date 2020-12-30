package common

import (
	"fmt"
	"runtime"
	"sync"
)

// parallelCommand is a type that declares functions
// the ps plugins can execute in parallel
type parallelCommand func(string) error

// parallelCommandResult is the result of a parallel
// command run
type parallelCommandResult struct {
	Name        string
	CommandName string
	Error       error
}

// RunCommandAgainstAllApps runs a given parallelCommand against all apps
func RunCommandAgainstAllApps(command parallelCommand, commandName string, parallelCount int) error {
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
		return RunCommandAgainstAllAppsSerially(command, commandName)
	}

	return RunCommandAgainstAllAppsInParallel(command, commandName, parallelCount)
}

// RunCommandAgainstAllAppsInParallel runs a given parallelCommand against all apps in parallel
func RunCommandAgainstAllAppsInParallel(command parallelCommand, commandName string, parallelCount int) error {
	apps, err := DokkuApps()
	if err != nil {
		LogWarn(err.Error())
		return nil
	}

	jobs := make(chan string, parallelCount)
	results := make(chan parallelCommandResult, len(apps))

	go allocateJobs(apps, jobs)
	done := make(chan error)
	go aggregateResults(results, done)
	createParallelWorkerPool(jobs, results, command, commandName, parallelCount)
	err = <-done

	return err
}

// RunCommandAgainstAllAppsSerially runs a given parallelCommand against all apps serially
func RunCommandAgainstAllAppsSerially(command parallelCommand, commandName string) error {
	apps, err := DokkuApps()
	if err != nil {
		LogWarn(err.Error())
		return nil
	}

	errorCount := 0
	for _, appName := range apps {
		LogInfo1(fmt.Sprintf("Running %s against app %s", commandName, appName))
		if err = command(appName); err != nil {
			LogWarn(fmt.Sprintf("Error running %s against app %s: %s", commandName, appName, err.Error()))
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

func aggregateResults(results chan parallelCommandResult, done chan error) {
	var parallelError error
	errorCount := 0
	for result := range results {
		if result.Error != nil {
			LogWarn(fmt.Sprintf("Error running %s against %s: %s", result.CommandName, result.Name, result.Error.Error()))
			errorCount++
		}
	}
	if errorCount > 0 {
		parallelError = fmt.Errorf("Encountered %d errors during parallel run", errorCount)
	}
	done <- parallelError
}

func createParallelWorker(jobs chan string, results chan parallelCommandResult, command parallelCommand, commandName string, wg *sync.WaitGroup, workerID int) {
	for job := range jobs {
		LogInfo1(fmt.Sprintf("Running command against %s", job))
		output := parallelCommandResult{
			Name:        job,
			CommandName: commandName,
			Error:       command(job),
		}
		results <- output
	}
	wg.Done()
}

func createParallelWorkerPool(jobs chan string, results chan parallelCommandResult, command parallelCommand, commandName string, numberOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go createParallelWorker(jobs, results, command, commandName, &wg, i)
	}
	wg.Wait()
	close(results)
}

type ReportFunc func(string) string

func CollectReport(appName string, flags map[string]ReportFunc) map[string]string {
	var sm sync.Map
	var wg sync.WaitGroup
	for flag, fn := range flags {
		wg.Add(1)
		go func(flag string, fn ReportFunc) {
			defer wg.Done()
			s := fn(appName)
			sm.Store(flag, s)
		}(flag, fn)
	}
	wg.Wait()

	infoFlags := map[string]string{}
	sm.Range(func(key interface{}, value interface{}) bool {
		infoFlags[key.(string)] = value.(string)
		return true
	})

	return infoFlags
}
