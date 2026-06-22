// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

// Entry is a single documentation page from the index.
type Entry struct {
	Title       string
	URL         string
	Description string
	Product     string
}

// Product is a documentation group from the index with its entry count.
type Product struct {
	Name  string
	Count int
}

// Index holds the parsed documentation catalog. It is safe for concurrent
// read access after construction.
type Index struct {
	Entries  []Entry
	products []Product
	idf      map[string]float64 // precomputed IDF weights for search ranking
}

// Products returns the product groups in index order.
func (idx *Index) Products() []Product {
	return idx.products
}

// EntryCount returns the total number of indexed entries.
func (idx *Index) EntryCount() int {
	return len(idx.Entries)
}
