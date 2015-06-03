package libxml2

/*
#cgo pkg-config: libxml-2.0
#include <stdbool.h>
#include "libxml/tree.h"
#include "libxml/parser.h"
#include "libxml/xpath.h"

// Macro wrapper function
static inline bool MY_xmlXPathNodeSetIsEmpty(xmlNodeSetPtr ptr) {
	return xmlXPathNodeSetIsEmpty(ptr);
}

// Because Go can't do pointer airthmetics...
static inline xmlNodePtr MY_xmlNodeSetTabAt(xmlNodePtr *nodes, int i) {
	return nodes[i];
}

// Change xmlIndentTreeOutput global, return old value, so caller can
// change it back to old value later
static inline int MY_setXmlIndentTreeOutput(int i) {
	int old = xmlIndentTreeOutput;
	xmlIndentTreeOutput = i;
	return old;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type XmlNodeType int

const (
	ElementNode XmlNodeType = iota + 1
	AttributeNode
	TextNode
	CDataSectionNode
	EntityRefNode
	EntityNode
	PiNode
	CommentNode
	DocumentNode
	DocumentTypeNode
	DocumentFragNode
	NotationNode
	HTMLDocumentNode
	DTDNode
	ElementDecl
	AttributeDecl
	EntityDecl
	NamespaceDecl
	XIncludeStart
	XIncludeEnd
	DocbDocumentNode
)

var _XmlNodeType_index = [...]uint8{0, 11, 24, 32, 48, 61, 71, 77, 88, 100, 116, 132, 144, 160, 167, 178, 191, 201, 214, 227, 238, 254}

const _XmlNodeType_name = `ElementNodeAttributeNodeTextNodeCDataSectionNodeEntityRefNodeEntityNodePiNodeCommentNodeDocumentNodeDocumentTypeNodeDocumentFragNodeNotationNodeHTMLDocumentNodeDTDNodeElementDeclAttributeDeclEntityDeclNamespaceDeclXIncludeStartXIncludeEndDocbDocumentNode`

func (i XmlNodeType) String() string {
	i -= 1
	if i < 0 || i+1 >= XmlNodeType(len(_XmlNodeType_index)) {
		return fmt.Sprintf("XmlNodeType(%d)", i+1)
	}
	return _XmlNodeType_name[_XmlNodeType_index[i]:_XmlNodeType_index[i+1]]
}

var ErrNodeNotFound = errors.New("node not found")
var ErrInvalidArgument = errors.New("invalid argument")

// Node defines the basic DOM interface
type Node interface {
	// pointer() returns the underlying C pointer. Only we are allowed to
	// slice it, dice it, do whatever the heck with it.
	pointer() unsafe.Pointer

	AddChild(Node)
	AppendChild(Node)
	ChildNodes() []Node
	OwnerDocument() *XmlDoc
	FindNodes(string) ([]Node, error)
	IsSameNode(Node) bool
	LastChild() Node
	NextSibling() Node
	NodeName() string
	NodeType() XmlNodeType
	ParetNode() Node
	PreviousSibling() Node
	SetNodeName(string)
	String() string
	TextContent() string
	ToString(int, bool) string
	Walk(func(Node) error)
}

type xmlNode struct {
	ptr *C.xmlNode
}

type XmlNode struct {
	*xmlNode
}

type XmlElement struct {
	*XmlNode
}

type XmlDoc struct {
	ptr  *C.xmlDoc
	root *C.xmlNode
}

type XmlText struct {
	*XmlNode
}

func wrapXmlElement(n *C.xmlElement) *XmlElement {
	return &XmlElement{wrapXmlNode((*C.xmlNode)(unsafe.Pointer(n)))}
}

func wrapXmlNode(n *C.xmlNode) *XmlNode {
	return &XmlNode{
		&xmlNode{
			ptr: (*C.xmlNode)(unsafe.Pointer(n)),
		},
	}
}

func wrapToNode(n *C.xmlNode) Node {
	switch XmlNodeType(n._type) {
	case ElementNode:
		return wrapXmlElement((*C.xmlElement)(unsafe.Pointer(n)))
	case TextNode:
		return &XmlText{&XmlNode{&xmlNode{ptr: n}}}
	default:
		return &XmlNode{&xmlNode{ptr: n}}
	}
}

func findNodes(n Node, xpath string) ([]Node, error) {
	ctx := C.xmlXPathNewContext((*C.xmlNode)(n.pointer()).doc)
	defer C.xmlXPathFreeContext(ctx)

	res := C.xmlXPathEvalExpression(stringToXmlChar(xpath), ctx)
	defer C.xmlXPathFreeObject(res)
	if C.MY_xmlXPathNodeSetIsEmpty(res.nodesetval) {
		return []Node(nil), nil
	}

	ret := make([]Node, res.nodesetval.nodeNr)
	for i := 0; i < int(res.nodesetval.nodeNr); i++ {
		ret[i] = wrapToNode(C.MY_xmlNodeSetTabAt(res.nodesetval.nodeTab, C.int(i)))
	}
	return ret, nil
}

func (n *xmlNode) pointer() unsafe.Pointer {
	return unsafe.Pointer(n.ptr)
}

func (n *xmlNode) AddChild(child Node) {
	C.xmlAddChild(n.ptr, (*C.xmlNode)(child.pointer()))
}

func (n *xmlNode) AppendChild(child Node) {
	// XXX There must be lots more checks here because AddChild does things
	// under the table like merging text nodes, freeing some nodes implicitly,
	// et al
	n.AddChild(child)
}

func (n *xmlNode) ChildNodes() []Node {
	return childNodes(n)
}

func wrapXmlDoc(n *C.xmlDoc) *XmlDoc {
	r := C.xmlDocGetRootElement(n) // XXX Should check for n == nil
	return &XmlDoc{ptr: n, root: r}
}

func (n *xmlNode) OwnerDocument() *XmlDoc {
	return wrapXmlDoc(n.ptr.doc)
}

func (n *xmlNode) FindNodes(xpath string) ([]Node, error) {
	return findNodes(n, xpath)
}

func (n *xmlNode) IsSameNode(other Node) bool {
	return n.pointer() == other.pointer()
}

func (n *xmlNode) LastChild() Node {
	return wrapToNode(n.ptr.last)
}

func (n *xmlNode) NodeName() string {
	return xmlCharToString(n.ptr.name)
}

func (n *xmlNode) NextSibling() Node {
	return wrapToNode(n.ptr.next)
}

func (n *xmlNode) ParetNode() Node {
	return wrapToNode(n.ptr.parent)
}

func (n *xmlNode) PreviousSibling() Node {
	return wrapToNode(n.ptr.prev)
}

func (n *xmlNode) SetNodeName(name string) {
	C.xmlNodeSetName(n.ptr, stringToXmlChar(name))
}

func (n *xmlNode) String() string {
	return n.ToString(0, false)
}

func (n *xmlNode) TextContent() string {
	return xmlCharToString(C.xmlXPathCastNodeToString(n.ptr))
}

func (n *xmlNode) ToString(format int, docencoding bool) string {
	buffer := C.xmlBufferCreate()
	defer C.xmlBufferFree(buffer)
	if format <= 0 {
		C.xmlNodeDump(buffer, n.ptr.doc, n.ptr, 0, 0)
	} else {
		oIndentTreeOutput := C.MY_setXmlIndentTreeOutput(1)
		C.xmlNodeDump(buffer, n.ptr.doc, n.ptr, 0, C.int(format))
		C.MY_setXmlIndentTreeOutput(oIndentTreeOutput)
	}
	return xmlCharToString(C.xmlBufferContent(buffer))
}

func (n *xmlNode) NodeType() XmlNodeType {
	return XmlNodeType(n.ptr._type)
}

func (n *xmlNode) Walk(fn func(Node) error) {
	panic("should not call walk on internal struct")
}

func (n *XmlNode) Walk(fn func(Node) error) {
	walk(n, fn)
}

func walk(n Node, fn func(Node) error) {
	if err := fn(n); err != nil {
		return
	}
	for _, c := range n.ChildNodes() {
		walk(c, fn)
	}
}

func childNodes(n Node) []Node {
	ret := []Node(nil)
	for chld := ((*C.xmlNode)(n.pointer())).children; chld != nil; chld = chld.next {
		ret = append(ret, wrapToNode(chld))
	}
	return ret
}

func NewDocument(version string) *XmlDoc {
	doc := C.xmlNewDoc(stringToXmlChar(version))
	return wrapXmlDoc(doc)
}

func (d *XmlDoc) pointer() unsafe.Pointer {
	return unsafe.Pointer(d.ptr)
}

func (d *XmlDoc) CreateElement(name string) *XmlElement {
	// XXX Should think about properly encoding the 'name'
	newNode := C.xmlNewNode(nil, stringToXmlChar(name))
	if newNode == nil {
		return nil
	}
	// XXX hmmm...
	newNode.doc = d.ptr
	return wrapXmlElement((*C.xmlElement)(unsafe.Pointer(newNode)))
}

func (d *XmlDoc) DocumentElement() Node {
	if d.ptr == nil || d.root == nil {
		return nil
	}

	return wrapToNode(d.root)
}

func (d *XmlDoc) FindNodes(xpath string) ([]Node, error) {
	root := d.DocumentElement()
	if root == nil {
		return nil, ErrNodeNotFound
	}
	return root.FindNodes(xpath)
}

func (d *XmlDoc) Encoding() string {
	return xmlCharToString(d.ptr.encoding)
}

func (d *XmlDoc) Free() {
	C.xmlFreeDoc(d.ptr)
	d.ptr = nil
	d.root = nil
}

func (d *XmlDoc) String() string {
	var xc *C.xmlChar
	i := C.int(0)
	C.xmlDocDumpMemory(d.ptr, &xc, &i)
	return xmlCharToString(xc)
}

func (d *XmlDoc) NodeType() XmlNodeType {
	return XmlNodeType(d.ptr._type)
}

func (d *XmlDoc) SetDocumentElement(n Node) {
	C.xmlDocSetRootElement(d.ptr, (*C.xmlNode)(n.pointer()))
}

func (n *XmlDoc) Walk(fn func(Node) error) {
	walk(wrapXmlNode(n.root), fn)
}

func (n *XmlText) Data() string {
	return xmlCharToString(n.ptr.content)
}

func (n *XmlText) Walk(fn func(Node) error) {
	walk(n, fn)
}
