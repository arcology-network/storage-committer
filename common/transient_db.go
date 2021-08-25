package common

type TransientDB struct {
	*DataStore
	parent DB
}

func NewTransientDB(parent DB) DB {
	return &TransientDB{
		DataStore: NewDataStore(),
		parent:    parent,
	}
}

func (db *TransientDB) Save(path string, v interface{}) {
	db.DataStore.Save(path, v)
}

func (db *TransientDB) Retrive(path string) interface{} {
	v := db.DataStore.Retrive(path)
	if v == nil {
		v = db.parent.Retrive(path)
	}
	return v
}

func (db *TransientDB) BatchSave(paths []string, states []interface{}) {
	db.DataStore.BatchSave(paths, states)
}

func (this *TransientDB) Print() {
	this.DataStore.Print()
}
