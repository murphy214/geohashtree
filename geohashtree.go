package geohashtree

/*
Flat, memory bound, k/v store based point-in-polygon alg. using geohash.
*/

import (
	"github.com/mmcloughlin/geohash"
	"math"
)

// Extrema structure (Bounding Box)
type Extrema struct {
	W float64
	E float64
	N float64
	S float64
}

// gets the extrema object of  a given geohash
func GetExtrema(ghash string) Extrema {
	box := geohash.BoundingBox(ghash)
	return Extrema{S: box.MinLat, W: box.MinLng, N: box.MaxLat, E: box.MaxLng}
}

// gets the extrema object of  a given geohash
func Middle(ghash string) []float64 {
	box := geohash.BoundingBox(ghash)
	lat, lng := box.Center()
	return []float64{lng, lat}
}

// gets a geohash
func Geohash(pt []float64, size int) string {
	return geohash.EncodeWithPrecision(pt[1], pt[0], uint(size))
}

// drill geohash
func ExpandGeohash(geohash string) []string {
	newlist := make([]string, 32)
	for pos, i := range "0123456789bcdefghjkmnpqrstuvwxyz" {
		newlist[pos] = geohash + string(i)
	}
	return newlist
}

// gets the starting geohashs from the bounds and minimum precision
func GetStartingHashs(bds Extrema, minp int) []string {
	en := Geohash([]float64{bds.E, bds.N}, minp)
	es := Geohash([]float64{bds.E, bds.S}, minp)
	wn := Geohash([]float64{bds.W, bds.N}, minp)
	ws := Geohash([]float64{bds.W, bds.S}, minp)

	// creating map and adding each corner geohash
	newmap := map[string]string{}
	newmap[en] = ""
	newmap[es] = ""
	newmap[wn] = ""
	newmap[ws] = ""

	// starting the iterating process
	bdsg := GetExtrema(en)
	deltax := bdsg.E - bdsg.W
	deltay := bdsg.N - bdsg.S
	for i := bds.W; i < bdsg.E; i += deltax {
		for j := bds.S; j < bdsg.N; j += deltay {
			ghash := Geohash([]float64{i, j}, minp)
			newmap[ghash] = ""
		}
	}
	newlist := make([]string, len(newmap))
	i := 0
	for k := range newmap {
		newlist[i] = k
		i++
	}
	return newlist
}

// The polygon structure used to build the index.
type Poly struct {
	Polygon  [][][]float64
	Extrema  Extrema                   // extrema of polygon
	Map      map[int]map[string]string // map of precision geohash fro every precision
	Min, Max int                       // minimum and maximum percentage
}

// creates the polygon structure
func CreatePolygon(polygon [][][]float64, minp, maxp int) *Poly {
	west, south, east, north := 180.0, 90.0, -180.0, -90.0
	maxpmap := map[string]string{}
	for _, cont := range polygon {
		for _, pt := range cont {
			x, y := pt[0], pt[1]
			// can only be one condition
			// using else if reduces one comparison
			if x < west {
				west = x
			} else if x > east {
				east = x
			}

			if y < south {
				south = y
			} else if y > north {
				north = y
			}
			ghash := Geohash(pt, maxp)
			maxpmap[ghash] = ""
		}
	}

	// logic for creating the smaller maps of each geohash
	currentmap := maxpmap
	totalmap := map[int]map[string]string{}
	totalmap[maxp] = maxpmap
	for i := maxp - 1; i >= minp; i-- {
		newmap := map[string]string{}
		for k := range currentmap {
			newmap[k[:len(k)-1]] = ""
		}
		totalmap[i] = newmap
		currentmap = newmap
	}

	return &Poly{
		Polygon: polygon,
		Extrema: Extrema{N: north, S: south, E: east, W: west},
		Map:     totalmap,
		Min:     minp,
		Max:     maxp,
	}
}

func (cont Poly) Pip(p []float64) bool {
	// Cast ray from p.x towards the right
	intersections := 0
	for _, c := range cont.Polygon {
		for i := range c {
			curr := c[i]
			ii := i + 1
			if ii == len(c) {
				ii = 0
			}
			next := c[ii]

			// Is the point out of the edge's bounding box?
			// bottom vertex is inclusive (belongs to edge), top vertex is
			// exclusive (not part of edge) -- i.e. p lies "slightly above
			// the ray"
			bottom, top := curr, next
			if bottom[1] > top[1] {
				bottom, top = top, bottom
			}
			if p[1] < bottom[1] || p[1] >= top[1] {
				continue
			}
			// Edge is from curr to next.

			if p[0] >= math.Max(curr[0], next[0]) ||
				next[1] == curr[1] {
				continue
			}

			// Find where the line intersects...
			xint := (p[1]-curr[1])*(next[0]-curr[0])/(next[1]-curr[1]) + curr[0]
			if curr[0] != next[0] && p[0] > xint {
				continue
			}

			intersections++
		}
	}

	return intersections%2 != 0
}

// gettting a hard point in polygon
func (polygon Poly) HardPip(geohash string) int {
	size := len(geohash)
	_, boolval := polygon.Map[size][geohash]
	if boolval {
		return 2
	}
	count := 0
	bds := GetExtrema(geohash)
	pts := [][]float64{{bds.W, bds.N}, {bds.E, bds.N}, {bds.E, bds.S}, {bds.W, bds.S}} // wn,en,es,ws
	for _, i := range pts {
		if polygon.Pip(i) {
			count += 1
		}
	}

	// if within
	if count == len(pts) {
		return 1
	} else if count > 0 {
		// part within
		return 2
	} else {
		// out
		return 3
	}

	return 3
}

// drills a single geohash to the maximum precision
func (poly *Poly) DrillGeohash(geohash string, newlist []string) []string {
	// creating channel
	c := make(chan []string)

	// checking to see if the geohash starting is within the polygon
	// covering a corner case
	if poly.Min == len(geohash) {
		hardpip := poly.HardPip(geohash)
		if hardpip == 1 {
			return append(newlist, geohash)
		}
	}

	// iterating through expanded geohash
	for _, newgeohash := range ExpandGeohash(geohash) {
		if len(newgeohash) <= poly.Max {
			hardpip := poly.HardPip(newgeohash)
			go func(newgeohash string, hardpip int, c chan []string) {
				if hardpip == 1 {
					c <- []string{newgeohash}
				} else if hardpip == 2 {
					c <- poly.DrillGeohash(newgeohash, []string{})
				} else {
					c <- []string{}
				}
			}(newgeohash, hardpip, c)
		} else {
			go func(c chan []string) {
				c <- []string{}
			}(c)
		}
	}

	for count := 0; count < 32; count++ {
		newlist = append(newlist, <-c...)
	}

	return newlist
}

// creates a polygon index given a polygon a min & max geohash precision
// returns a string with all geohashs that are within the polygon.
func MakePolygonIndex(polygon [][][]float64, minp, maxp int) []string {
	poly := CreatePolygon(polygon, minp, maxp)
	// getting staritng geohashs
	s_geohashs := GetStartingHashs(poly.Extrema, minp)

	// iterating through starting geohashs
	c := make(chan []string)
	for _, ghash := range s_geohashs {
		go func(ghash string, c chan []string) {
			c <- poly.DrillGeohash(ghash, []string{})
		}(ghash, c)
	}
	total := []string{}
	for range s_geohashs {
		total = append(total, <-c...)
	}

	return total
}
