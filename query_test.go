package geohashtree

import (
	"math/rand"
	"testing"
)

var testtree, _ = OpenGeohashTreeCSV("test_data/a.csv")

//var testtree2, _ = OpenGeohashTreeBoltDB("../county.db")

var testtree2, _ = OpenGeohashTreeBoltDB("test_data/a.db")
var keys = getkeys("test_data/a.csv")

func getkeys(filename string) []string {
	scanner, _ := NewScannerFile(filename)
	keys := []string{}
	for scanner.Next() {
		key, _ := scanner.KeyValue()
		keys = append(keys, key)
	}
	return keys
}

func BenchmarkQueryMap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testtree.Query(RandomPt())
	}
}

func BenchmarkQueryBoltDB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testtree2.Query(RandomPt())
	}
}

func BenchmarkGetMap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testtree.Get(keys[rand.Intn(len(keys)-1)])
	}
}

func BenchmarkGetBoltDB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testtree2.Get(keys[rand.Intn(len(keys)-1)])
	}
}
