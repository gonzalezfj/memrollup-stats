package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Frequency  float64
	OutputFile string
	JSONOutput bool
	Verbose    bool
	Command    string
	PID        int
	StartTime  time.Time
}

var cfg = &Config{
	Frequency: 1.0,
}

// Get returns the current configuration.
func Get() *Config {
	return cfg
}

// ParseArgs parses command line arguments.
func ParseArgs() error {
	flag.Float64Var(&cfg.Frequency, "F", 1.0, "Sampling frequency in Hz")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output file for statistics")
	flag.BoolVar(&cfg.JSONOutput, "j", false, "Output in JSON format")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose mode")
	showHelp := flag.Bool("h", false, "Show help message")
	showVersion := flag.Bool("V", false, "Show version")

	flag.Parse()

	if *showHelp {
		ShowHelpMessage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("memrollup-stats version %s\n", Version)
		os.Exit(0)
	}

	// Get remaining arguments as command
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("no command specified. Use -h for help")
	}
	cfg.Command = strings.Join(args, " ")

	if cfg.Frequency <= 0 {
		return fmt.Errorf("frequency must be a positive number")
	}

	if cfg.Frequency > 100 {
		fmt.Printf("Warning: High frequency (%.1f Hz) may impact system performance\n", cfg.Frequency)
	}

	return nil
}

// ValidateEnvironment checks if the environment meets the requirements.
func ValidateEnvironment() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("this program requires Linux")
	}

	// Check if /proc is mounted and accessible
	if _, err := os.Stat("/proc/self/smaps_rollup"); os.IsNotExist(err) {
		return fmt.Errorf("smaps_rollup not available in /proc. This program requires Linux kernel 4.14 or later")
	}

	if cfg.OutputFile != "" {
		dir := filepath.Dir(cfg.OutputFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", dir)
		}

		// Check write permissions
		if file, err := os.OpenFile(filepath.Join(dir, ".write_test"), os.O_WRONLY|os.O_CREATE, 0666); err != nil {
			return fmt.Errorf("cannot write to directory: %s", dir)
		} else {
			file.Close()
			os.Remove(filepath.Join(dir, ".write_test"))
		}
	}

	return nil
}

// Version is the current version of the application.
var Version = "2.0.0"

// ShowHelpMessage displays the help message.
func ShowHelpMessage() {
	fmt.Println(`memrollup-stats - Memory usage monitoring and statistics tool

Usage:
  memrollup-stats -F <frequency_hz> [-o output_file] [-j] [-v] command [args...]

Options:
  -F <frequency_hz>    Sampling frequency in Hz (default: 1)
  -o <output_file>     Output file for statistics (default: stdout)
  -j                   Output in JSON format
  -v                   Verbose mode
  -h                   Show this help message
  -V                   Show version

Examples:
  # Monitor a process with 10Hz frequency
  memrollup-stats -F 10 ./my_program

  # Save statistics to CSV file
  memrollup-stats -F 2 -o stats.csv ./my_program

  # Output JSON for data analysis
  memrollup-stats -F 1 -j -o stats.json ./my_program

Exit codes:
  0  - Success
  1  - General error
  2  - Invalid arguments
  3  - Process start failure
  4  - Permission denied
  5  - System resource unavailable

Author: Facundo Gonzalez <facujgg@gmail.com>
License: MIT
Version: 2.0.0`)
}
