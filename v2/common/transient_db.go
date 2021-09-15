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
		// if v != nil {
		// 	v = v.(TypeInterface).Deepcopy()
		// }
	}
	return v
}

func (db *TransientDB) BatchSave(paths []string, dict interface{}) {
	db.DataStore.BatchSave(paths, dict)
}

func (this *TransientDB) Print() {
	this.DataStore.Print()
}
