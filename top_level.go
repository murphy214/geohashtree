package geohashtree

import (
	"github.com/boltdb/bolt"
	"math"
	"math/rand"
	"strconv"
)

var Bds = Extrema{N: 55.0, S: 10.0, W: -130.0, E: -70.0}

type GeohashTree struct {
	Type   string // current either map or boltdb
	Map    map[string]string
	Min    int
	Max    int
	Dummy  string
	Bucket *bolt.Bucket
	BoltDB *bolt.DB
}

// get function for different db types
func (tree *GeohashTree) Get(key string) (string, bool) {
	if tree.Type == "map" {
		val, boolval := tree.Map[key]
		return val, boolval
	} else if tree.Type == "boltdb" {
		val := string(tree.Bucket.Get([]byte(key)))
		return val, len(val) > 0
	}

	return "", false
}

// queries the entire flat index
func (tree *GeohashTree) Query(point []float64) (string, bool) {
	ghash := Geohash(point, tree.Max)
	for i := tree.Min; i <= tree.Max; i++ {
		val, boolval := tree.Get(ghash[:i])
		if !boolval {
			return val, false
		} else if val != "-1" {
			return val, true
		}
	}
	return "", false
}

// opens a csv tree
func OpenGeohashTreeCSV(filename string) (*GeohashTree, error) {
	mymap := ReadFileMap(filename)
	dummy := mymap["dummy"]
	minstr := mymap["min"]
	maxstr := mymap["max"]
	mind, err := strconv.ParseInt(minstr, 10, 64)
	if err != nil {
		return &GeohashTree{}, err
	}
	maxd, err := strconv.ParseInt(maxstr, 10, 64)
	if err != nil {
		return &GeohashTree{}, err
	}
	return &GeohashTree{
		Type:  "map",
		Map:   mymap,
		Min:   int(mind),
		Max:   int(maxd),
		Dummy: dummy,
	}, err
}

// opens a boltdb tree
func OpenGeohashTreeBoltDB(filename string) (*GeohashTree, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return &GeohashTree{}, err
	}
	tx, err := db.Begin(false)
	if err != nil {
		return &GeohashTree{}, err
	}
	buck := tx.Bucket([]byte("world"))

	// creating tree
	tree := &GeohashTree{
		Type:   "boltdb",
		BoltDB: db,
		Bucket: buck,
	}
	dummy, _ := tree.Get("dummy")
	minstr, _ := tree.Get("min")
	maxstr, _ := tree.Get("max")
	mind, err := strconv.ParseInt(minstr, 10, 64)
	if err != nil {
		return &GeohashTree{}, err
	}
	maxd, err := strconv.ParseInt(maxstr, 10, 64)
	if err != nil {
		return &GeohashTree{}, err
	}
	tree.Max = int(maxd)
	tree.Min = int(mind)
	tree.Dummy = dummy
	return tree, err
}

// get a random point
func RandomPt() []float64 {
	deltax := math.Abs(Bds.W - Bds.E)
	deltay := math.Abs(Bds.N - Bds.S)
	return []float64{(rand.Float64() * deltax) + Bds.W, (rand.Float64() * deltay) + Bds.S}
}
