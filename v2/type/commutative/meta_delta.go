package commutative

type MetaDelta struct {
	added   []string // added keys in the current block
	removed []string // removed keys in the current block
}

func (this *MetaDelta) Added() interface{}   { return this.added }
func (this *MetaDelta) Removed() interface{} { return this.removed }
