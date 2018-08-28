package geohashtree

import (
	"testing"
)

var testtree, _ = OpenGeohashTreeCSV("test_data/a.csv")

func BenchmarkQuery(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testtree.Query(RandomPt())
	}
}
