package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Csv struct {
	file   *os.File
	reader *bufio.Reader
	header []string
	sepCh  string
}

type InvalidCsvError struct {
	cause error
}

func (e InvalidCsvError) Error() string {
	return fmt.Sprintf("invalid csv: %s", e.cause.Error())
}

type CsvEofError struct{}

func (e CsvEofError) Error() string {
	return "eof"
}

func openCsv(filepath string, commaCh string) (Csv, error) {
	csv := Csv{}
	f, err := os.Open(filepath)
	if err != nil {
		return csv, err
	}

	csv.reader = bufio.NewReader(f)

	hdrLn, err := csv.reader.ReadString('\n')
	if err != nil {
		f.Close()
		return csv, InvalidCsvError{cause: err}
	}

	hdrLn = strings.TrimSuffix(hdrLn, "\n")

	csv.header = strings.Split(hdrLn, commaCh)

	if len(csv.header) == 0 {
		f.Close()
		return csv, InvalidCsvError{errors.New("no columns")}
	}

	csv.sepCh = commaCh
	csv.file = f

	return csv, nil
}

type CsvEntry struct {
	columnIndex int
	columnName  string
	value       string
}

func (csv Csv) readNextLine() ([]CsvEntry, error) {
	entryLn, err := csv.reader.ReadString('\n')
	if err != nil {
		return nil, CsvEofError{}
	}

	entryLn = strings.TrimSuffix(entryLn, "\n")
	colsVals := strings.Split(entryLn, csv.sepCh)
	//fmt.Println(colsVals)
	if len(colsVals) != len(csv.header) {
		return nil, InvalidCsvError{cause: errors.New("invalid num of cols")}
	}

	var b bytes.Buffer
	ents := []CsvEntry{}
	for idx, elem := range colsVals {
		b.WriteString(elem)
		b.WriteString(" ")
		ents = append(
			ents,
			CsvEntry{
				columnIndex: idx,
				columnName:  csv.header[idx],
				value:       elem,
			})
	}
	b.WriteString("\n")
	print(b.String())

	return ents, nil
}

func (csv Csv) close() {
	csv.file.Close()
}
