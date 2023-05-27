package indexer

type Conflicted struct {
	Unsuccessful
}

func (Conflicted) Refund() uint64 { return 0 }
