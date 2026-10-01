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
