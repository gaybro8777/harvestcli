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
)

type associateCommand struct {
	search *os.File
	click  *os.File
	output *os.File

	timestampIndex  int
	appIndex        int
	indexIndex      int
	queryIDIndex    int
	userIndex       int
	contextIndex    int
	queryIndex      int
	queryParamIndex int

	clickQueryIDIndex int
}

func (c *associateCommand) parse(arguments []string) {
	flagSet := flag.NewFlagSet("associate", flag.ExitOnError)
	search := flagSet.String("s", "searches.csv", "Search source")
	click := flagSet.String("c", "clicks.csv", "Click source")
	output := flagSet.String("o", "associated.csv", "Output file")
	flagSet.IntVar(&c.timestampIndex, "it", 0, "Index of timestamp column in search CSV")
	flagSet.IntVar(&c.appIndex, "ia", 1, "Index of appID column in search CSV")
	flagSet.IntVar(&c.indexIndex, "ii", 2, "Index of indexName column in search CSV")
	flagSet.IntVar(&c.queryIDIndex, "iqid", 3, "Index of queryID column in search CSV")
	flagSet.IntVar(&c.userIndex, "iu", 4, "Index of userID column in search CSV")
	flagSet.IntVar(&c.contextIndex, "ic", 5, "Index of context column in search CSV")
	flagSet.IntVar(&c.queryIndex, "iq", 6, "Index of query column in search CSV")
	flagSet.IntVar(&c.queryParamIndex, "iqp", 7, "Index of queryParam column in search CSV")
	flagSet.IntVar(&c.clickQueryIDIndex, "icqid", 2, "Index of queryID column in click CSV")

	flagSet.Parse(arguments)

	searchFile, err := os.Open(*search)
	if err != nil {
		log.Fatalf("Could not open search file %s: %s", *search, err)
	}

	clickFile, err := os.Open(*click)
	if err != nil {
		log.Fatalf("Could not open click file %s: %s", *click, err)
	}

	outputFile, err := os.OpenFile(*output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Could not create output file %s: %s", *output, err)
	}

	fmt.Printf("Processing clicks from %s and searches from %s into %s...\n", *click, *search, *output)
	c.search = searchFile
	c.click = clickFile
	c.output = outputFile
}

func (c *associateCommand) Close() {
	defer c.search.Close()
	defer c.click.Close()
	defer c.output.Close()
}

func (c *associateCommand) Run() {
	c.addClickEvent(c.getClickSet())
}

func (c *associateCommand) getClickSet() map[string]bool {
	fmt.Printf("Generating click set...\n")
	csvr := csv.NewReader(c.click)
	result := make(map[string]bool)
	for {
		record, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error while reading click file: %s", err)
		}
		result[record[c.clickQueryIDIndex]] = true
	}
	fmt.Printf("Got %d unique queries with clicks\n", len(result))
	return result
}

func (c *associateCommand) addClickEvent(set map[string]bool) {
	fmt.Printf("Creating search set...\n")
	csvr := csv.NewReader(c.search)
	linecount := 0
	for {
		record, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error while reading search file: %s", err)
		}
		_, hasClick := set[record[c.queryIDIndex]]

		data := []string{
			record[c.timestampIndex],
			record[c.appIndex],
			record[c.indexIndex],
			record[c.queryIDIndex],
			record[c.userIndex],
			record[c.contextIndex],
			strconv.FormatBool(hasClick),
			record[c.queryIndex],      // Last one so even if it has comma we can still use command line tools like sort or cut with , as the separator
			record[c.queryParamIndex], // Last one so even if it has comma we can still use command line tools like sort or cut with , as the separator
		}
		utils.WriteCSV(c.output, data)
		linecount++
	}
	fmt.Printf("Processed %d searches\n", linecount)
}

func NewAssociateCommand(arguments []string) *associateCommand {
	cmd := associateCommand{}
	cmd.parse(arguments)
	return &cmd
}
