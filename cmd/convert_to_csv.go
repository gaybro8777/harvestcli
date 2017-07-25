package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	event "github.com/algolia/harvestcli/event"
	"github.com/algolia/harvestcli/utils"
)

type convertToCSVCommand struct {
	input  *os.File
	output *os.File
}

func (c *convertToCSVCommand) parse(arguments []string) {
	flagSet := flag.NewFlagSet("convert", flag.ExitOnError)
	input := flagSet.String("i", "input.json", "Input file")
	output := flagSet.String("o", "output.csv", "Output file")
	flagSet.Parse(arguments)

	inputFile, err := os.Open(*input)
	if err != nil {
		log.Fatalf("Could not open input file %s: %s", *input, err)
	}

	outputFile, err := os.OpenFile(*output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Could not create output file %s: %s", *output, err)
	}

	fmt.Printf("Converting %s into %s...\n", *input, *output)
	c.input = inputFile
	c.output = outputFile
}

func (c *convertToCSVCommand) Close() {
	defer c.input.Close()
	defer c.output.Close()
}

func (c *convertToCSVCommand) Run() {
	scanner := bufio.NewScanner(c.input)
	for scanner.Scan() {
		c.convertLine(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Could not read source file: %s", err)
	}
}

func (c *convertToCSVCommand) convertLine(line string) {
	var logLine event.LogLine
	if err := json.Unmarshal([]byte(line), &logLine); err != nil {
		log.Fatalf("Could not unmarshal log line: %s\n", err)
	}

	switch event.GetLogType(logLine) {
	case event.SEARCH:
		c.convertSearch(logLine.JSONPayload)
	case event.CLICK:
		c.convertClick(logLine.JSONPayload)
	default:
		log.Fatal("Unknown log type")
	}
}

func (c *convertToCSVCommand) convertSearch(payload event.Payload) {
	data := []string{
		payload.Timestamp.String(),
		payload.AppID,
		payload.Index,
		payload.QueryID,
		payload.UserID,
		payload.Context,
		payload.Query,
		payload.QueryParameters,
	}
	utils.WriteCSV(c.output, data)
}

func (c *convertToCSVCommand) convertClick(payload event.Payload) {
	data := []string{
		payload.Timestamp.String(),
		payload.AppID,
		payload.QueryID,
		payload.Position.String(),
		payload.ObjectID,
	}
	utils.WriteCSV(c.output, data)
}

func NewConvertToCSVCommand(arguments []string) *convertToCSVCommand {
	cmd := convertToCSVCommand{}
	cmd.parse(arguments)
	return &cmd
}
