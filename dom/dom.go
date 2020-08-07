package dom

import (
	"runtime"
	"sync"

	"github.com/474420502/libxml2/xpath"
)

var docPool sync.Pool

func init() {
	SetupXPathCallback()
	docPool = sync.Pool{}
	docPool.New = func() interface{} {
		return Document{}
	}
}

func SetupXPathCallback() {
	xpath.WrapNodeFunc = WrapNode
}

func WrapDocument(n uintptr) *Document {
	doc := docPool.Get().(Document)
	// doc := &Document{}
	doc.mortal = false
	doc.ptr = n

	runtime.SetFinalizer(&doc, func(obj interface{}) bool {
		obj.(*Document).AutoFree()
		return true
	})

	return &doc
}
