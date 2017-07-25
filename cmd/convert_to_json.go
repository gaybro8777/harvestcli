package cmd

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	event "github.com/algolia/harvestcli/event"
	"github.com/algolia/harvestcli/utils"
)

type convertToJSONCommand struct {
	input  *os.File
	output *os.File
}

func (c *convertToJSONCommand) parse(arguments []string) {
	flagSet := flag.NewFlagSet("convert", flag.ExitOnError)
	input := flagSet.String("i", "input.csc", "Input file")
	output := flagSet.String("o", "output.json", "Output file")
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

func (c *convertToJSONCommand) Close() {
	defer c.input.Close()
	defer c.output.Close()
}

func (c *convertToJSONCommand) Run() {
	r := csv.NewReader(c.input)
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		json := c.convertLine(line)
		utils.WriteJSON(c.output, json)
	}
}

func (c *convertToJSONCommand) convertLine(line []string) event.SearchEvent {
	return event.SearchEvent{
		Timestamp: json.Number(line[0]),
		Index:     line[2],
		AppID:     line[1],
		QueryID:   line[3],
		UserID:    line[4],
		Context:   line[5],
		Query:     line[7],
		// QueryParameters: line[8],
	}
}

func NewConvertToJSONCommand(arguments []string) *convertToJSONCommand {
	cmd := convertToJSONCommand{}
	cmd.parse(arguments)
	return &cmd
}
