package bot

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/quintilesims/iqvbot/db"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func newMemoryStore(t *testing.T) *db.MemoryStore {
	store := db.NewMemoryStore()
	if err := db.Init(store); err != nil {
		t.Fatal(err)
	}

	return store
}
