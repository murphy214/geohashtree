package geohashtree

/*
This files handles abstracting the core function geohashtree.go uses.
*/

import (
	"fmt"
	"github.com/paulmach/go.geojson"
	"io/ioutil"
	"os"
	"strings"
)

var Increment = 2 // change this if you want the increment to increase

// creates a string that can be appended to a csv file
func CleanOutput(outputgeohashs []string, idstring string, minval int) string {
	// creating stringlist
	mymap := map[string]string{}
	newlist := []string{}
	for i, val := range outputgeohashs {
		currentval := val
		for len(currentval) != minval {
			_, boolval := mymap[currentval[:len(currentval)-1]]
			if !boolval {
				mymap[currentval[:len(currentval)-1]] = ""
				newlist = append(newlist, fmt.Sprintf("%s,%s", currentval[:len(currentval)-1], "-1"))
			}
			currentval = currentval[:len(currentval)-1]
		}
		outputgeohashs[i] = fmt.Sprintf("%s,%s", val, idstring)
	}
	return strings.Join(append(newlist, outputgeohashs...), "\n") + "\n"
}

type IndexOutput struct {
	Min, Max      int
	File          *os.File
	FileName      string
	TotalPolygons int
}

// creates a csv to start appending to
func CreateCSV(filename string, minp, maxp int) (*IndexOutput, error) {
	os.Create(filename)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	file.WriteString(fmt.Sprintf("GEOHASH,ID\nmin,%d\nmax,%d\ndummy,-1\n", minp, maxp))
	return &IndexOutput{
		Min:           minp,
		Max:           maxp,
		File:          file,
		FileName:      filename,
		TotalPolygons: 0,
	}, err
}

func (output *IndexOutput) AddFeature(feature *geojson.Feature, field string) string {
	val, boolval := feature.Properties[field]
	var valstr string
	if !boolval {
		return ""
	} else {
		valstr, boolval = val.(string)
		if !boolval {
			return ""
		}
	}
	output.TotalPolygons++
	if feature.Geometry.Type == "Polygon" {
		return CleanOutput(
			MakePolygonIndex(feature.Geometry.Polygon, output.Min, output.Max),
			valstr,
			output.Min,
		)
	} else if feature.Geometry.Type == "MultiPolygon" {
		totaloutput := []string{}
		for _, polygon := range feature.Geometry.MultiPolygon {
			totaloutput = append(totaloutput, MakePolygonIndex(polygon, output.Min, output.Max)...)
		}
		return CleanOutput(totaloutput, valstr, output.Min)
	}
	return ""
}

// creates an index from geojson and dumps it into a csv
func IndexFromGeoJSON(filename string, outfilename string, minp, maxp int, geojsonfield string) error {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	fc, err := geojson.UnmarshalFeatureCollection(bs)
	if err != nil {
		return err
	}

	output, err := CreateCSV(outfilename, minp, maxp)
	if err != nil {
		return err
	}
	current := 0
	for current < len(fc.Features) {
		nextcurrent := current + Increment
		if nextcurrent > len(fc.Features) {
			nextcurrent = len(fc.Features)
		}
		c := make(chan string, nextcurrent-current)
		for _, feature := range fc.Features[current:nextcurrent] {
			go func(feature *geojson.Feature, c chan string) {
				c <- output.AddFeature(feature, geojsonfield)
			}(feature, c)
		}
		for i := current; i < nextcurrent; i++ {
			val := <-c
			output.File.WriteString(val)
		}
		current = nextcurrent
		fmt.Printf("\r[%d/%d] of features written to output csv.", current, len(fc.Features))
	}
	fmt.Printf("\nFinished making output csv: %s\n", outfilename)
	return err
}
