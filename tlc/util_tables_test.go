package tlc

import "testing"

func TestVectRemoveAllPreservesJavaRawElementCount(t *testing.T) {
	v := NewVectWithCapacity[string](4)
	v.AddElement("a")
	v.AddElement("b")
	v.AddElement("c")

	v.RemoveAll(1)
	if got := v.String(); got != "{a}" {
		t.Fatalf("after shrink String = %q, want {a}", got)
	}

	v.RemoveAll(3)
	if got := v.String(); got != "{a,b,c}" {
		t.Fatalf("after re-expanding within capacity String = %q, want stale Java backing elements", got)
	}
}

func TestObjLongTableKeysEnumeratorIsLiveLikeJava(t *testing.T) {
	table := NewObjLongTable[string](4)
	keys := table.Keys()
	table.Put("later", 17)

	got, ok := keys.NextElement()
	if !ok || got != "later" {
		t.Fatalf("live ObjLongTable keys returned %q/%v, want later/true", got, ok)
	}
	if _, ok := keys.NextElement(); ok {
		t.Fatalf("live ObjLongTable keys returned unexpected second element")
	}
}

func TestSemanticNodeLongTableKeysEnumeratorIsLiveLikeJava(t *testing.T) {
	table := NewSemanticNodeLongTable(4)
	keys := table.Keys()
	node := &SemanticNodeBase{Image: "later"}
	table.Put(node, 17)

	if got := keys.NextElement(); got != node {
		t.Fatalf("live SemanticNodeLongTable keys returned %v, want inserted node", got)
	}
	if got := keys.NextElement(); got != nil {
		t.Fatalf("live SemanticNodeLongTable keys returned unexpected second node %v", got)
	}
}

func TestLongObjTableNilElementIsJavaNullSentinel(t *testing.T) {
	table := NewLongObjTable[*BEGraphNode](5)
	loc := table.Put(11, nil)
	if table.Size() != 1 {
		t.Fatalf("after nil put size = %d, want Java count 1", table.Size())
	}
	if got, ok := table.Get(11); ok || got != nil {
		t.Fatalf("nil element lookup = %v/%v, want Java null miss", got, ok)
	}

	node := NewBEGraphNode(11)
	loc2 := table.Put(11, node)
	if loc2 != loc {
		t.Fatalf("non-nil replacement location = %d, want original nil slot %d", loc2, loc)
	}
	if table.Size() != 2 {
		t.Fatalf("after replacing null-sentinel slot size = %d, want Java count 2", table.Size())
	}
	if got, ok := table.Get(11); !ok || got != node {
		t.Fatalf("non-nil lookup = %v/%v, want original node", got, ok)
	}

	table.Put(22, NewBEGraphNode(22))
	if table.Size() != 2 {
		t.Fatalf("after grow size = %d, want nil sentinel dropped during Java rehash", table.Size())
	}
}
