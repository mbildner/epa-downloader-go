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
	"time"
)

const BASE_PATH = "chemicals"
const DEFAULT_SECONDS = 2
const MAX_RUNNING = 10

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
	timeToWait := flag.Int("interval", DEFAULT_SECONDS, "seconds to wait between downloading next pdf")

	flag.Parse()

	seed := loadSeedJSON(*seedLocation)

	for _, chemical := range seed {
		time.Sleep(time.Duration(*timeToWait) * time.Second)
		fmt.Printf("Download: %s (%d of %d)\n", chemical.Name)
		err := getChemicalFromSeed(chemical)
		if err != nil {
			fmt.Printf("failure: %s", chemical.Name)
		}
	}
}
