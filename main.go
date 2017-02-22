package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

const BASE_PATH = "chemicals"
const DEFAULT_MAX_CONCURRENT = 4

type downloadSeed struct {
	Id   string
	HREF string
	Name string
}

func loadSeedJSON(seedLocation string) []downloadSeed {
	file, err := ioutil.ReadFile(seedLocation)
	if err != nil {
		fmt.Println("failed to read file")
		os.Exit(1)
	}

	seed := []downloadSeed{}

	json.Unmarshal(file, &seed)
	return seed
}

func ensureBaseDirectory() {
	if _, err := os.Stat(BASE_PATH); os.IsNotExist(err) {
		os.Mkdir(BASE_PATH, 0700)
	}
}

func getChemicalFromSeed(seed downloadSeed) error {
	fileTarget := path.Join(BASE_PATH, seed.Name+".pdf")
	out, err := os.Create(fileTarget)
	if err != nil {
		return err
	}
	defer out.Close()

	doc, _ := goquery.NewDocument(seed.HREF)

	summaryLink := doc.Find("a[href$=\"summary.pdf\"]").First()
	summaryURL, _ := summaryLink.Attr("href")

	err = downloadPDFToFile(summaryURL, out)
	if err != nil {
		return err
	}

	return nil
}

func downloadPDFToFile(url string, out *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	io.Copy(out, resp.Body)
	return nil
}

func main() {
	ensureBaseDirectory()
	seedLocation := flag.String("seed", "seed.json", "location for download seed json file")
	maxConcurrentDownloads := flag.Int("max-concurrent", DEFAULT_MAX_CONCURRENT, "maximum number of files to download at the same time")

	flag.Parse()

	if *maxConcurrentDownloads < 1 {
		fmt.Println("minimum concurrent downloads is defaulting to 1\n")
		*maxConcurrentDownloads = 1
	}

	fmt.Printf("maximum concurrent downloads set to: %d\n", *maxConcurrentDownloads)

	backpressure := make(chan bool, *maxConcurrentDownloads)

	seed := loadSeedJSON(*seedLocation)

	for _, chemical := range seed {
		backpressure <- true
		go func(chemical downloadSeed, backpressure chan bool) {
			err := getChemicalFromSeed(chemical)
			if err != nil {
				fmt.Printf("could not download: %s\n", chemical.Name)
			}
			<-backpressure
		}(chemical, backpressure)
	}
}
