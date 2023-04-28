package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	maxDepth          = 10
	successMsgFormat  = "Successfully composed file: %s"
	fileReadError     = "Error reading included file %s: %v"
	includeProcessErr = "Error processing included file %s: %v"
)

var inputFile string

var rootCmd = &cobra.Command{
	Use:   "ycl",
	Short: "ycl is a CLI tool for composition YAML.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(inputFile)
	},
}

type Command interface {
	Execute() error
}

type command struct{}

func New() Command {
	return &command{}
}

func (c *command) Execute() error {
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input YAML file path")
	rootCmd.MarkFlagRequired("input")
	return rootCmd.Execute()
}

func run(inputPath string) error {
	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	inputString := string(inputBytes)
	inputString, err = processIncludes(inputString, filepath.Dir(inputPath), 0)
	if err != nil {
		return err
	}

	var data yaml.Node
	err = yaml.Unmarshal([]byte(inputString), &data)
	if err != nil {
		return err
	}

	outputBytes, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	outputPath := inputPath[:len(inputPath)-len(filepath.Ext(inputPath))] + ".mod" + filepath.Ext(inputPath)
	err = os.WriteFile(outputPath, outputBytes, 0644)
	if err != nil {
		return err
	}
	log.Info().Msgf(successMsgFormat, outputPath)

	return nil
}

func processIncludes(input string, baseDir string, depth int) (string, error) {
	if depth > maxDepth {
		return "", errors.New("maximum include depth exceeded")
	}

	includePattern := regexp.MustCompile(`((?m)^\s*)#include:\s*(.+)`)
	output := includePattern.ReplaceAllStringFunc(input, func(match string) string {
		matches := includePattern.FindStringSubmatch(match)
		indent := matches[1]
		filePath := filepath.Join(baseDir, matches[2])

		fileContentBytes, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal().Err(err).Msgf(fileReadError, filePath, err)
			return ""
		}

		fileContent := string(fileContentBytes)
		fileContent, err = processIncludes(fileContent, filepath.Dir(filePath), depth+1)
		if err != nil {
			log.Fatal().Err(err).Msgf(includeProcessErr, filePath, err)
			return ""
		}

		fileContentIndented := addIndent(fileContent, indent)
		return fileContentIndented
	})

	return strings.TrimSpace(output), nil
}

func addIndent(input, indent string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
