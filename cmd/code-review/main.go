package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"

	"github.com/lmquang/code-review/pkg/diff"
	"github.com/lmquang/code-review/pkg/git"
	"github.com/lmquang/code-review/pkg/gpt"
)

type Config struct {
	OpenAIAPIKey string `yaml:"openai_api_key"`
	OpenAIModel  string `yaml:"openai_model"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: code-review <command> [<args>]")
		fmt.Println("Commands:")
		fmt.Println(" set    Set the OpenAI API Key and/or model")
		fmt.Println(" review Run the code review process")
		return
	}

	switch os.Args[1] {
	case "set", "s":
		handleSetCommand()
	case "review", "r":
		handleReviewCommand()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleSetCommand() {
	setCmd := flag.NewFlagSet("set", flag.ExitOnError)
	openAIAPIKey := setCmd.String("openai-api-key", "", "Set the OpenAI API Key")
	openAIModel := setCmd.String("openai-model", "", "Set the OpenAI Model")

	err := setCmd.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("Error parsing set command: %v", err)
	}

	if *openAIAPIKey == "" && *openAIModel == "" {
		log.Fatal("Please provide at least one of -openai-api-key or -openai-model")
	}

	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading existing config: %v", err)
	}

	if *openAIAPIKey != "" {
		config.OpenAIAPIKey = *openAIAPIKey
	}
	if *openAIModel != "" {
		config.OpenAIModel = *openAIModel
	}

	if err := saveConfig(config); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Println("Configuration has been saved successfully.")
}

func handleReviewCommand() {
	reviewCmd := flag.NewFlagSet("review", flag.ExitOnError)
	ignoreFlag := reviewCmd.String("ignore", "", "Comma-separated list of files or extensions to ignore (e.g., '*.yaml,*.json,docs.go')")

	err := reviewCmd.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("Error parsing review command: %v", err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	config := parseConfig(*ignoreFlag)

	gitClient := git.NewClient()
	diffFormatter := diff.NewFormatter(diff.SplitAndTrimPatterns(*ignoreFlag))
	gptClient := gpt.NewOpenAIClient(config.OpenAIAPIKey)
	if config.OpenAIModel != "" {
		gptClient.Client().SetModel(config.OpenAIModel)
	}

	diff, changedFiles, err := gitClient.GetDiff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}

	if diff == "" {
		fmt.Println("No changes detected in the current branch.")
		return
	}

	originalContent, formattedDiff, errors := diffFormatter.Format(diff, changedFiles)
	if len(errors) > 0 {
		fmt.Println("Encountered errors while processing some files:")
		for _, err := range errors {
			fmt.Printf("- %v\n", err)
		}
		fmt.Println("Continuing with the files that were processed successfully.")
	}

	if formattedDiff == "" {
		fmt.Println("No changes to review after applying ignore patterns.")
		return
	}

	gptResponse, err := gptClient.Review(originalContent, formattedDiff)
	if err != nil {
		log.Fatalf("Error sending to GPT: %v", err)
	}

	fmt.Println("GPT Review:")
	fmt.Println(gptResponse)
}

func parseConfig(ignoreFlag string) Config {
	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
	}

	if config.OpenAIAPIKey == "" {
		config.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	}

	if config.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set. Please set it using 'code-review set -openai-api-key YOUR_API_KEY' or as an environment variable.")
	}

	return config
}

func saveConfig(config Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".code-review.yaml")
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config to YAML: %w", err)
	}

	return os.WriteFile(configPath, yamlData, 0600)
}

func loadConfig() (Config, error) {
	var config Config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("error getting home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".code-review.yaml")
	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshaling YAML: %w", err)
	}

	return config, nil
}
