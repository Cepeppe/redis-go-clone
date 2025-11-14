package main

/*
	TODO EXTENSION: Hasing based rdb persistence

	Apply hasing on keys to determine their bucket.
	Use multiple bucket files.
	Use an array of dirty boolean flags (one for each bucket)
	Apply rdb persistence to dirty bucket keys only
	--> possibility to have light weight persistence overhead

*/

type MemoryDataDirty struct {
	// isDirty []bool
	// lastRdbSnapshotAt []int64    //millis timestamps
}
