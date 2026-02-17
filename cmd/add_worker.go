/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"fmt"

	"github.com/padiazg/hexago/internal/generator"
	"github.com/spf13/cobra"
)

var (
	workerType     string
	workerInterval string
	workerWorkers  int
	workerQueueSize int
)

// addWorkerCmd represents the add worker command
var addWorkerCmd = &cobra.Command{
	Use:   "worker <name>",
	Short: "Add a background worker",
	Long: `Add a background worker for processing async tasks using goroutines and channels.

Worker types:
  queue     - Queue-based worker (default) - processes jobs from a channel
  periodic  - Periodic worker - runs tasks at intervals
  event     - Event-driven worker - reacts to events

Workers include:
  - Goroutine-based concurrent processing
  - Channel-based communication
  - Graceful shutdown with context
  - WaitGroup coordination
  - Start/Stop lifecycle methods

Example:
  hexago add worker EmailWorker --type queue
  hexago add worker HealthCheckWorker --type periodic --interval 1m
  hexago add worker NotificationWorker --type event`,
	Args: cobra.ExactArgs(1),
	RunE: runAddWorker,
}

func init() {
	addCmd.AddCommand(addWorkerCmd)

	addWorkerCmd.Flags().StringVarP(&workerType, "type", "t", "queue", "Worker type (queue|periodic|event)")
	addWorkerCmd.Flags().StringVar(&workerInterval, "interval", "5m", "Interval for periodic workers (e.g., 5m, 1h)")
	addWorkerCmd.Flags().IntVar(&workerWorkers, "workers", 5, "Number of concurrent workers for queue type")
	addWorkerCmd.Flags().IntVar(&workerQueueSize, "queue-size", 100, "Queue size for queue-based workers")
}

func runAddWorker(cmd *cobra.Command, args []string) error {
	workerName := args[0]

	if err := validateComponentName(workerName); err != nil {
		return err
	}

	// Validate worker type
	validTypes := map[string]bool{
		"queue":    true,
		"periodic": true,
		"event":    true,
	}

	if !validTypes[workerType] {
		return fmt.Errorf("invalid worker type '%s'. Valid types: queue, periodic, event", workerType)
	}

	config, err := generator.GetCurrentProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	fmt.Printf("📦 Adding worker: %s (%s)\n", workerName, workerType)
	fmt.Printf("   Project: %s\n", config.ProjectName)

	if workerType == "periodic" {
		fmt.Printf("   Interval: %s\n", workerInterval)
	} else if workerType == "queue" {
		fmt.Printf("   Workers: %d\n", workerWorkers)
		fmt.Printf("   Queue size: %d\n", workerQueueSize)
	}
	fmt.Println()

	// Generate worker
	workerConfig := generator.WorkerConfig{
		Type:      workerType,
		Interval:  workerInterval,
		Workers:   workerWorkers,
		QueueSize: workerQueueSize,
	}

	gen := generator.NewWorkerGenerator(config)
	if err := gen.Generate(workerName, workerConfig); err != nil {
		return fmt.Errorf("failed to generate worker: %w", err)
	}

	fmt.Println("\n✅ Worker added successfully!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Implement the worker logic in the process method\n")
	fmt.Printf("  2. Register the worker in cmd/run.go:\n")
	fmt.Printf("     - Create worker instance\n")
	fmt.Printf("     - Add to worker manager\n")
	fmt.Printf("     - Start with context\n")
	fmt.Printf("  3. Test the worker with unit tests\n")

	if workerType == "queue" {
		fmt.Printf("  4. Submit jobs to the worker:\n")
		fmt.Printf("     worker.Submit(Job{...})\n")
	}

	return nil
}
