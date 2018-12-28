package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/ajiyoshi-vg/goberkeleydb/bdb"
)

// Data represents a key-value data
type Data map[string]string

func main() {
	var jsonFile string
	var dbFile string

	flag.StringVar(&jsonFile, "input-from", "data/sample-data.json", "read JSON from")
	flag.StringVar(&dbFile, "output-to", "data/sample-bdb.db", "write database to")
	flag.Parse()

	writeDB(jsonFile, dbFile)
}

func writeDB(jsonFile, dbFile string) {
	var data Data

	b, err := ioutil.ReadFile(jsonFile)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, &data)

	if err != nil {
		panic(err)
	}

	db, err := bdb.OpenBDB(bdb.NoEnv, bdb.NoTxn, dbFile, nil, bdb.BTree, bdb.DbCreate, 0)

	if err != nil {
		panic(err)
	}

	for k, v := range data {
		if err = db.Put(bdb.NoTxn, []byte(k), []byte(v), 0); err != nil {
			panic(err)
		}
	}

	db.Close(0)
}
