package cmd

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/algolia/harvestcli/utils"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type mergeCommand struct {
	search *os.File
	output *os.File

	timestampIndex int
	indexIndex     int
	userIndex      int
	queryIndex     int
	clickIndex     int

	lineCount  int
	mergeCount int
	skipCount  int

	debug bool
}

func (c *mergeCommand) parse(arguments []string) {
	flagSet := flag.NewFlagSet("merge", flag.ExitOnError)
	search := flagSet.String("s", "searches.csv", "Search source with clicks data")
	output := flagSet.String("o", "output.csv", "Output file")
	flagSet.IntVar(&c.timestampIndex, "it", 0, "Index of timestamp column in CSV")
	flagSet.IntVar(&c.indexIndex, "ii", 2, "Index of indexName column in CSV")
	flagSet.IntVar(&c.userIndex, "iu", 4, "Index of userID column in CSV")
	flagSet.IntVar(&c.queryIndex, "iq", 7, "Index of query column in CSV")
	flagSet.IntVar(&c.clickIndex, "ic", 6, "Index of click column in CSV")
	flagSet.BoolVar(&c.debug, "d", false, "Print debug statements")
	flagSet.Parse(arguments)

	searchFile, err := os.Open(*search)
	if err != nil {
		log.Fatalf("Could not open search file %s: %s", *search, err)
	}

	outputFile, err := os.OpenFile(*output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Could not create output file %s: %s", *output, err)
	}

	fmt.Printf("Processing searches from %s into %s...\n", *search, *output)
	c.search = searchFile
	c.output = outputFile
}

func (c *mergeCommand) Close() {
	defer c.search.Close()
	defer c.output.Close()
}

func (c *mergeCommand) Run() {
	fmt.Printf("Merging searches...\n")
	csvr := csv.NewReader(c.search)
	line, err := c.readLine(csvr)
	if err != nil {
		fmt.Println("No search queries to process.")
		return
	}

	for {
		nextLine, err := c.readLine(csvr)
		shouldContinue := err == nil && nextLine != nil

		isTerminal := c.isTerminal(line, nextLine)

		if isTerminal {
			err = utils.WriteCSV(c.output, line)
			if err != nil {
				log.Fatalf("Could not write terminal search %s\n", err)
			}
			c.mergeCount++
		}

		if !shouldContinue {
			break
		}
		line = nextLine
	}
	fmt.Printf("Processed %d searches, got %d terminal searches, skipped %d\n", c.lineCount, c.mergeCount, c.skipCount)
}

func (c *mergeCommand) isTerminal(line []string, nextLine []string) bool {
	c.d(fmt.Sprintf("Processing %s...", line[c.queryIndex]))
	if c.hasClick(line) {
		c.d(fmt.Sprintf("\thas a click: TERMINAL\n"))
		return true
	}

	// No next query? terminal
	if len(nextLine) == 0 || nextLine == nil {
		c.d(fmt.Sprintf("\tis last query: TERMINAL\n"))
		return true
	}

	// Not same user? terminal
	if line[c.userIndex] != nextLine[c.userIndex] {
		c.d(fmt.Sprintf("\tis followed by a query belonging to someone else: TERMINAL\n"))
		return true
	}

	// Not same index? terminal
	if line[c.indexIndex] != nextLine[c.indexIndex] {
		c.d(fmt.Sprintf("\tis followed by a query for another index: TERMINAL\n"))
		return true
	}

	// if queries are within less than 200ms, not terminal
	nextLineTime, _ := strconv.ParseInt(nextLine[c.timestampIndex], 10, 64)
	lineTime, _ := strconv.ParseInt(line[c.timestampIndex], 10, 64)
	if nextLineTime-lineTime < 200 {
		c.d(fmt.Sprintf("\tnext query is close enough: NOT TERMINAL\n"))
		return false
	}

	// Edit distance > 1 ? terminal
	if levenshtein.DistanceForStrings([]rune(line[c.queryIndex]), []rune(nextLine[c.queryIndex]), levenshtein.DefaultOptions) > 2 {
		c.d(fmt.Sprintf("\tedit distance with next query is > 2: TERMINAL\n"))
		return true
	}

	// not terminal
	c.d(fmt.Sprintf("\tdefault: NOT TERMINAL\n"))
	return false
}

func (c *mergeCommand) hasClick(line []string) bool {
	hasClick, _ := strconv.ParseBool(line[c.clickIndex])
	return hasClick
}

func (c *mergeCommand) readLine(csvr *csv.Reader) ([]string, error) {
	record, err := csvr.Read()
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		c.skipCount++
		log.Printf("Error while reading search file: %s", err)
		return nil, err
	}
	c.lineCount++
	return record, nil
}

func (c *mergeCommand) d(log string) {
	if c.debug {
		fmt.Print(log)
	}
}

func NewMergeCommand(arguments []string) *mergeCommand {
	cmd := mergeCommand{}
	cmd.parse(arguments)
	return &cmd
}
