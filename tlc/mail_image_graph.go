/*
 * Copyright (c) 2000, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 2000, 2021, Oracle and/or its affiliates. All rights reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.  Oracle designates this
 * particular file as subject to the "Classpath" exception as provided
 * by Oracle in the LICENSE file that accompanied this code.
 *
 * This code is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * version 2 for more details (a copy is included in the LICENSE file that
 * accompanied this code).
 *
 * You should have received a copy of the GNU General Public License version
 * 2 along with this work; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * Please contact Oracle, 500 Oracle Parkway, Redwood Shores, CA 94065 USA
 * or visit www.oracle.com if you need additional information or have any
 * questions.
 */

// OpenJDK PartiallyOrderedSet, DigraphNode and PartialOrderIterator operations.
// Hash iteration supplies the tie order; cycles and their blocked successors
// are omitted. Iterators copy in-degrees but read current out-edges on next().
package tlc

import "sync/atomic"

type mailImageDigraphNode struct {
	data              *MailImageSPI
	outNodes, inNodes mailObjectHashMap
	inDegree          int32
	identityHash      atomic.Int32
}

func (n *mailImageDigraphNode) mailObjectHashCode() int32                 { return n.mailObjectIdentityHash() }
func (n *mailImageDigraphNode) mailObjectEquals(k mailObjectHashKey) bool { return n == k }
func (n *mailImageDigraphNode) mailObjectClass() *MailActivationClass     { return nil } // DigraphNode is not Comparable.
func (n *mailImageDigraphNode) mailObjectClassName() string               { return "javax.imageio.spi.DigraphNode" }
func (n *mailImageDigraphNode) mailObjectIdentityHash() int32 {
	if cb := DefaultMailImageIOEnvironment.GraphNodeIdentityHash; cb != nil {
		return cb(n)
	}
	if h := n.identityHash.Load(); h != 0 {
		return h
	}
	h := mailFlavorIdentityHash.Add(1) & 0x7fffffff
	if h == 0 {
		h = 1
	}
	n.identityHash.CompareAndSwap(0, h)
	return n.identityHash.Load()
}
func (n *mailImageDigraphNode) mailObjectComparableClass() *MailActivationClass { return nil }
func (n *mailImageDigraphNode) mailObjectCompare(mailObjectHashKey) int32 {
	panic(NewClassCastException())
}
func (n *mailImageDigraphNode) addEdge(node *mailImageDigraphNode) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.outNodes.contains(node) {
		return false
	}
	n.outNodes.put(node, true)
	if node == nil {
		panic(NewNullPointerException())
	}
	node.inNodes.put(n, true)
	node.inDegree++
	return true
}
func (n *mailImageDigraphNode) hasEdge(node *mailImageDigraphNode) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	return n.outNodes.contains(node)
}
func (n *mailImageDigraphNode) removeEdge(node *mailImageDigraphNode) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	if !n.outNodes.contains(node) {
		return false
	}
	n.outNodes.remove(node)
	if node == nil {
		panic(NewNullPointerException())
	}
	node.inNodes.remove(n)
	node.inDegree--
	return true
}
func mailImageNodeArray(m *mailObjectHashMap) []*mailImageDigraphNode {
	a := make([]*mailImageDigraphNode, 0, int(m.size))
	i := m.iterator()
	for i.hasNext() {
		k := i.nextNode().objectKey
		if k == nil {
			a = append(a, nil)
		} else {
			a = append(a, k.(*mailImageDigraphNode))
		}
	}
	return a
}
func (n *mailImageDigraphNode) dispose() {
	if n == nil {
		panic(NewNullPointerException())
	}
	in := mailImageNodeArray(&n.inNodes)
	for _, node := range in {
		node.removeEdge(n)
	}
	out := mailImageNodeArray(&n.outNodes)
	for _, node := range out {
		n.removeEdge(node)
	}
}

type mailImagePartiallyOrderedSet struct{ nodes mailObjectHashMap }

func (p *mailImagePartiallyOrderedSet) contains(data *MailImageSPI) bool {
	return p.nodes.contains(data)
}
func (p *mailImagePartiallyOrderedSet) add(data *MailImageSPI) bool {
	if p.nodes.contains(data) {
		return false
	}
	node := &mailImageDigraphNode{data: data}
	p.nodes.put(data, node)
	return true
}
func (p *mailImagePartiallyOrderedSet) remove(data *MailImageSPI) bool {
	v := p.nodes.get(data)
	if v == nil {
		return false
	}
	p.nodes.remove(data)
	v.(*mailImageDigraphNode).dispose()
	return true
}
func (p *mailImagePartiallyOrderedSet) clear() { p.nodes.clear() }
func (p *mailImagePartiallyOrderedSet) node(data *MailImageSPI) *mailImageDigraphNode {
	v := p.nodes.get(data)
	if v == nil {
		return nil
	}
	return v.(*mailImageDigraphNode)
}
func (p *mailImagePartiallyOrderedSet) setOrdering(first, second *MailImageSPI) bool {
	a, b := p.node(first), p.node(second)
	b.removeEdge(a)
	return a.addEdge(b)
}
func (p *mailImagePartiallyOrderedSet) unsetOrdering(first, second *MailImageSPI) bool {
	a, b := p.node(first), p.node(second)
	return a.removeEdge(b) || b.removeEdge(a)
}
func (p *mailImagePartiallyOrderedSet) hasOrdering(first, second *MailImageSPI) bool {
	return p.node(first).hasEdge(p.node(second))
}
func (p *mailImagePartiallyOrderedSet) iterator() *mailImagePartialOrderIterator {
	i := &mailImagePartialOrderIterator{}
	source := p.nodes.iterator()
	for source.hasNext() {
		node := source.nextNode().objectValue.(*mailImageDigraphNode)
		degree := node.inDegree
		i.inDegrees.put(node, degree)
		if degree == 0 {
			i.zeroList = append(i.zeroList, node)
		}
	}
	return i
}

type mailImagePartialOrderIterator struct {
	zeroList  []*mailImageDigraphNode
	inDegrees mailObjectHashMap
}

func (i *mailImagePartialOrderIterator) hasNext() bool { return len(i.zeroList) != 0 }
func (i *mailImagePartialOrderIterator) next() *MailImageSPI {
	if len(i.zeroList) == 0 {
		panic(NewNoSuchElementException())
	}
	first := i.zeroList[0]
	i.zeroList = i.zeroList[1:]
	out := first.outNodes.iterator()
	for out.hasNext() {
		node := out.nextNode().objectKey.(*mailImageDigraphNode)
		v := i.inDegrees.get(node)
		if v == nil {
			panic(NewNullPointerException())
		}
		degree := v.(int32) - 1
		i.inDegrees.put(node, degree)
		if degree == 0 {
			i.zeroList = append(i.zeroList, node)
		}
	}
	return first.data
}
func (i *mailImagePartialOrderIterator) remove() { panic(NewUnsupportedOperationException()) }

// Class identities for the hardwired reader/writer providers. The common
// hierarchy is metadata; full SPI constructors and registry lifecycle are
// separate required ports.
var mailImageRegisterableServiceClass = &MailActivationClass{Name: "javax.imageio.spi.RegisterableService"}
var mailImageIIOServiceProviderClass = &MailActivationClass{Name: "javax.imageio.spi.IIOServiceProvider", Parents: []*MailActivationClass{mailFlavorObjectClass, mailImageRegisterableServiceClass}}
var mailImageReaderWriterSPIClass = &MailActivationClass{Name: "javax.imageio.spi.ImageReaderWriterSpi", Parents: []*MailActivationClass{mailImageIIOServiceProviderClass}}
var mailImageReaderSPIClass = &MailActivationClass{Name: "javax.imageio.spi.ImageReaderSpi", Parents: []*MailActivationClass{mailImageReaderWriterSPIClass}}
var mailImageWriterSPIClass = &MailActivationClass{Name: "javax.imageio.spi.ImageWriterSpi", Parents: []*MailActivationClass{mailImageReaderWriterSPIClass}}
var mailStandardImageReaderClasses = mailStandardImageClasses("Reader", mailImageReaderSPIClass)
var mailStandardImageWriterClasses = mailStandardImageClasses("Writer", mailImageWriterSPIClass)

func mailStandardImageClasses(direction string, parent *MailActivationClass) []*MailActivationClass {
	result := make([]*MailActivationClass, 0, 6)
	for _, kind := range []string{"gif.GIF", "bmp.BMP", "wbmp.WBMP", "tiff.TIFF", "png.PNG", "jpeg.JPEG"} {
		result = append(result, &MailActivationClass{Name: "com.sun.imageio.plugins." + kind + "Image" + direction + "Spi", Parents: []*MailActivationClass{parent}})
	}
	return result
}
