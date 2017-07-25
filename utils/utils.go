package utils

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func WriteJSON(file *os.File, payload interface{}) error {
	json, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	json = append(json, "\n"...)
	_, err = file.Write(json)
	return err
}

func WriteCSV(file *os.File, line []string) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(line); err != nil {
		return err
	}
	return nil
}

func GetCSVLine(line []string) (string, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	defer writer.Flush()

	err := writer.Write(line)
	if err != nil {
		return "", err
	}

	writer.Flush()
	return buffer.String(), nil
}

func AppendToFile(filename string, line string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		return err
	}
	return nil
}

func ExecCommand(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func EnsureEmptyDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
