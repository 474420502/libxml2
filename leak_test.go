package libxml2

import (
	"testing"

	"github.com/474420502/libxml2/types"
)

func Leak() {
	doc, err := ParseHTMLString(`<html><head></head><body><a href="https://www.google.com"></a></body></html>`)
	defer doc.Free()
	if err != nil {
		panic(err)
	}
	result, err := doc.Find("//a")
	if err != nil {
		panic(err)
	}
	iter := result.NodeIter()
	if iter.Next() {
		if attr, err := iter.Node().(types.Element).GetAttribute("href"); err != nil || attr.Value() != "https://www.google.com" {
			panic(err)
		} else {
			// log.Println(attr.Value())
		}
	}

}

func BenchmarkLeak(t *testing.B) {
	for i := 0; i < 5000000; i++ {
		Leak()
	}

	// time.Sleep(time.Second * 100)
}
