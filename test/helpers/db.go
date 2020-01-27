package helpers

import "d3/orm"

type dbAdapterWithQueryCounter struct {
	queryCounter int
	dbAdapter orm.StorageAdapter
}

func NewDbAdapterWithQueryCounter(dbAdapter orm.StorageAdapter) *dbAdapterWithQueryCounter {
	wrappedAdapter := &dbAdapterWithQueryCounter{dbAdapter: dbAdapter}

	dbAdapter.AfterQuery(func(_ string, _ ...interface{}) {
		wrappedAdapter.queryCounter++
	})

	return wrappedAdapter
}

func (d *dbAdapterWithQueryCounter) QueryCounter() int {
	return d.queryCounter
}

func (d *dbAdapterWithQueryCounter) DbAdapter() orm.StorageAdapter {
	return d.dbAdapter
}



