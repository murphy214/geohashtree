package geohashtree

// this file is for reading / putting key / values in a map or db.

import (
	"bufio"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// function that gets linecount of csv file
func linecount(filename string) (int, error) {
	lc := int64(0)
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		lc++
	}
	return int(lc), s.Err()
}

// sorts csv for use
func SortCSV(filename string) error {
	//cmd := exec.Command("sort", "-k", "1n", "-r", "-S", "5G", filename)
	os.Remove("tmp.csv")
	cmd := "(head -n 2 a.csv && tail -n +3 a.csv | sort -k 1n -S 5g) >> tmp.csv"
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err
	}

	err = os.Remove(filename)
	if err != nil {
		return err
	}
	err = os.Rename("tmp.csv", filename)
	if err != nil {
		return err
	}
	return err
}

func SplitRow(val string) (string, string) {
	if strings.Contains(val, ",") {
		vals := strings.Split(val, ",")
		return vals[0], vals[1]
	}
	return "", ""
}

// the scanner absraction below enables
// apis to be done quickly for different dbs while using
// boilerplate code other than the put method.
type CSVScanner struct {
	Scanner *bufio.Scanner
}

func (scanner *CSVScanner) Next() bool {
	return scanner.Scanner.Scan()
}

func (scanner *CSVScanner) KeyValue() (string, string) {
	return SplitRow(scanner.Scanner.Text())
}

// opens the output csv file as a CSVScanner.
func NewScannerFile(filename string) (*CSVScanner, error) {
	file, err := os.Open(filename)
	if err != nil {
		return &CSVScanner{}, err
	}
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	return &CSVScanner{Scanner: scanner}, err
}

// this function reads the csv file into a pure go map
func ReadFileMap(filename string) map[string]string {
	scanner, _ := NewScannerFile(filename)
	mymap := map[string]string{}
	for scanner.Next() {
		key, val := scanner.KeyValue()
		mymap[key] = val
	}
	return mymap
}

// creates a bolt db database from a flat csv file
func CreateBoltDB(filename string, outfilename string) error {
	// sorting given csv file
	err := SortCSV(filename)
	if err != nil {
		return err
	}

	// geting line count
	lc, err := linecount(filename)
	if err != nil {
		return err
	}

	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open(outfilename, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// opening scanner file
	scanner, _ := NewScannerFile(filename)
	newlist := [][]string{}

	updatedb := func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("world"))
		if err != nil {
			return err
		}

		// Adding all key value stores from csv
		for _, row := range newlist {
			err = bucket.Put([]byte(row[0]), []byte(row[1]))
			if err != nil {
				return err
			}

		}
		return nil
	}

	// scanning through each k/v
	currentline := 0
	for scanner.Next() {
		k, v := scanner.KeyValue()
		newlist = append(newlist, []string{k, v})
		// updating db when 100k k/v's are buffered.
		if len(newlist) == 100000 {
			fmt.Printf("\r[%d/%d] Inserting k/v into %s...", currentline, lc, outfilename)
			err := db.Update(updatedb)
			if err != nil {
				return err
			}
			newlist = [][]string{}
		}
		currentline++
	}
	// adding the final remaining newlist if needed
	if len(newlist) > 0 {
		fmt.Printf("\r[%d/%d] Inserting k/v into %s...", currentline, lc, outfilename)
		err := db.Update(updatedb)
		if err != nil {
			return err
		}

	}
	return err
}
