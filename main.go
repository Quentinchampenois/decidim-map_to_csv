package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var ioReader io.Reader

	if len(os.Args) > 1 {
		url := os.Args[1]
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		ioReader = resp.Body
	} else {
		fmt.Println("Working in development mode. Using fixture.html file.")
		fmt.Println("To run in production mode, use: decidim-map_to_csv https://your-decidim-instance.com/your-page-with-map")

		file, err := os.Open("fixture.html")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		ioReader = file
	}

	doc, err := goquery.NewDocumentFromReader(ioReader)
	if err != nil {
		log.Fatal(err)
	}

	// From Map ID element
	// Get data-markers-data attribute content
	dataMarkersData, found := doc.Find("#map").Attr("data-markers-data")
	if !found {
		log.Fatal("No Decidim map found in page")
	}

	var markers []map[string]interface{}
	err = json.Unmarshal([]byte(dataMarkersData), &markers)
	if err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("map-geocoding-%s.csv", time.Now().Format("2006-01-02-15-04-05"))
	fmt.Printf("Generating '%s' file...", filename)
	csvFile, err := os.Create(filename)
	if err != nil {
		log.Fatal("Error creating CSV file:", err)
	}
	defer csvFile.Close()

	// Create a CSV writer
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write header to the CSV file
	header := []string{"title", "address", "latitude", "longitude", "url"}
	if err := writer.Write(header); err != nil {
		log.Fatal("Error writing CSV header:", err)
	}

	// Write data to the CSV file
	for _, record := range markers {
		rowData := []string{
			strings.TrimSpace(record["title"].(string)),
			strings.TrimSpace(record["address"].(string)),
			fmt.Sprintf("%f", record["latitude"].(float64)),
			fmt.Sprintf("%f", record["longitude"].(float64)),
			strings.TrimSpace(record["link"].(string)),
		}
		if err := writer.Write(rowData); err != nil {
			log.Fatal("Error writing CSV record:", err)
		}
	}

}
