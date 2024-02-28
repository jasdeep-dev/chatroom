package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type DataRecord interface{}

func BackupData(d DataRecord, dataFile string) {
	jsonData, err := json.Marshal(d)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Open the file in append mode
	file, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Append the JSON data to the file
	if _, err := file.Write(jsonData); err != nil {
		fmt.Println("Error appending JSON to file:", err)
		return
	}
	if _, err := file.WriteString("\n"); err != nil {
		fmt.Println("Error appending newline to file:", err)
		return
	}
}

type DataCollection interface {
	Restore([]byte)
}

func RestoreData(ds DataCollection, dataFile string) {
	file, err := os.Open(dataFile) // replace with your file name
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read the file line by line
	for scanner.Scan() {
		ds.Restore(scanner.Bytes())

	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
