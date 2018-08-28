# geohashtree

### What is it? 

This library is designed to construct and use large k/v stores to do super quick point-in-polygon queries. This isn't something that I haven't tried before in go but this time with a much cleaner index generation implementation. (which is all thats done currently) I still have to write the db interface that can work with some of the most used k/v stores. The actual point-in-polygon algorithm from here is dead simple. 

### Example Usage

This example shows the index generation for a given polygon. The size of the actual index is huge so here we just print the first 100.

```golang
package main

import (
	"fmt"
	"github.com/murphy214/geohashtree"
	"github.com/paulmach/go.geojson"
)

func main() {
	vals := `{"geometry": {"type": "Polygon", "coordinates": [[[-97.94860839843749, 42.44778143462245], [-97.97607421875, 42.65820178455667], [-98.5308837890625, 42.44372793752476], [-98.8714599609375, 42.06560675405716], [-98.85498046875, 42.459940352216556], [-98.9813232421875, 42.342305278572816], [-98.997802734375, 41.795888098191426], [-98.953857421875, 41.35619553438905], [-98.5968017578125, 41.7180304600481], [-98.624267578125, 41.95949009892467], [-98.173828125, 42.204107493733176], [-97.9705810546875, 42.020732852644294], [-98.349609375, 41.89001042401827], [-98.118896484375, 41.84910468610387], [-97.833251953125, 41.857287927691345], [-97.72338867187499, 42.248851700720955], [-97.22351074218749, 42.49235259142821], [-97.09167480468749, 42.0615286181226], [-97.020263671875, 42.62183364891663], [-97.94860839843749, 42.44778143462245]]]}, "type": "Feature", "properties": {}}`
	feature, _ := geojson.UnmarshalFeature([]byte(vals))

	// making a geohash tree of min size 5 and max 9 geohash
	geohashs := geohashtree.MakePolygonIndex(feature.Geometry.Polygon, 5, 9)

	for _, ghash := range geohashs[:100] {
		fmt.Println(ghash)
	}
}
```

**If we were to the put the output geohashs on a map it would look like this**

![](https://user-images.githubusercontent.com/10904982/44731827-03d52100-aab2-11e8-81f9-afee22344a66.png)
![](https://user-images.githubusercontent.com/10904982/44731828-03d52100-aab2-11e8-94e7-7bad8cb5e288.png)
