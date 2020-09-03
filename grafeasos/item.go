package grafeasos

import (
	grafeaspb "github.com/grafeas/client-go/0.1.0"
)

const noNewOcc = "no new occurrences generated"

// Item implements a MetadataItem.
type Item struct {
	Occurrence *grafeaspb.V1beta1Occurrence // The Occurrence this Item wraps.
}

// Name returns the name of the group of Item.
func (item *Item) Name() string {
	return item.Occurrence.NoteName
}

// String returns a string version of this Item.
func (item *Item) String() string {
	if nil != item.Occurrence {
		return item.Occurrence.Name + ";" + item.Occurrence.NoteName
	}

	return noNewOcc
}
