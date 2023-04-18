// MIT License
//
// Copyright (c) 2021 Xiantu Li
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package json

import (
	"fmt"
	"unsafe"
)

type tireTreeNode struct {
	next int16 // 下一个状态
	idx  int16 // 只有 '"' 才是借宿标志，才有 idx
}
type tireTree struct {
	tree [][128]tireTreeNode // 状态
	// ptree *[1 << 20]tireTreeNode
	tags []*TagInfo
}

func initTireTreeNode(tree *[128]tireTreeNode) {
	for i := range tree {
		tree[i].idx = -1
		tree[i].next = -1
	}
}
func NewTireTree(tags []*TagInfo) (root *tireTree, err error) {
	for i, tag := range tags {
		tags[i].TagName = tag.TagName + `"`
	}
	root = &tireTree{
		tree: make([][128]tireTreeNode, 1, 4),
		tags: tags,
	}
	initTireTreeNode(&root.tree[0])

out:
	for idx, tag := range tags {
		key := tag.TagName
		status := &root.tree[0]
		for iKey, c := range []byte(key) {
			k := c % 128
			n := &status[k]

			// 没有被占领或是叶子结点
			if n.next < 0 {
				// 没有被占领
				if n.idx < 0 {
					//占领此叶节点
					n.idx = int16(idx)
					continue out
				}
				// 叶子节点
				old := root.tags[n.idx].TagName
				if old == key {
					err = fmt.Errorf("duplicate key: %s", key)
					return
				}

				// 已经是 old 的终点 '"' 了。
				if len(old) == iKey+1 {
					err = fmt.Errorf("error key: %s", key)
					return
				}
				// 修改老的 status
				nOld := *n
				n.idx = -1
				n.next = int16(len(root.tree))

				// 给旧的 node 添加状态
				root.tree = append(root.tree, [128]tireTreeNode{})
				status = &root.tree[len(root.tree)-1] // 创建新的状态
				initTireTreeNode(status)

				kOld := old[iKey+1] % 128
				status[kOld] = nOld

				// kNew := key[iKey+1] % 128
				// if kNew != kOld {
				// 	//占领此叶节点
				// 	status[k].idx = int16(idx + 1)
				// 	continue out
				// }

				// key 的 next 在 for 的下一轮再处理！
				continue
			}
			status = &root.tree[n.next]
		}
	}

	if cap(root.tree) > len(root.tree) {
		// tree := make([][128]tireTreeNode, 0, len(root.tree))
		// tree = append(tree, root.tree...)
		// root.tree = tree
		root.tree = root.tree[:len(root.tree):len(root.tree)]
	}

	// root.ptree = (*[1 << 20]tireTreeNode)(unsafe.Pointer(&root.tree[0]))

	// 处理：合并同类树，避免树太高

	return
}

func (root *tireTree) zipTree() (rennew bool) {
	for current := range root.tree {
		if current == len(root.tree)-1 {
			break
		}
		m := make(map[uint8]struct{})
		for j := range root.tree[current] {
			if root.tree[current][j].next >= 0 {
				m[uint8(j)] = struct{}{}
			}
		}

		// TODO: 实际上如果和 current 冲突，还可以放其他节点去，一样可以压缩；不过都只能本 status 一起放
		for j := range root.tree[current] {
			if next := root.tree[current][j].next; next >= 0 {
				status := &root.tree[next]
				canZip := true
				for j := range status {
					if nextNest := status[j].next; nextNest >= 0 {
						if _, ok := m[uint8(nextNest)]; ok {
							// 和父节点有冲突，不适合压缩节点;
							canZip = false
						}
					}
				}
				if !canZip {
					continue
				}

				//即将删除的节点
				statusDeleted := next

				// 开始处理压缩逻辑
				for j := range status {
					if nextNest := status[j].next; nextNest >= 0 {
						root.tree[current][j] = status[j]
					}
				}

				root.tree = append(root.tree[:statusDeleted], root.tree[statusDeleted:]...)

				// 开始修正复制后的状态
				for i := range root.tree {
					for j := range root.tree[i] {
						if root.tree[i][j].next == statusDeleted {
							// root.tree[i][j].next = int16(i)
							root.tree[i][j].next = int16(current)
							continue
						}
						if root.tree[i][j].next > statusDeleted {
							root.tree[i][j].next--
						}
					}
				}

				// 还需再次探测是否可以继续压缩
				rennew = true
				return
			}
		}

	}
	return
}

func (root *tireTree) Get(key string) *TagInfo {
	p := (*[1 << 20]tireTreeNode)(unsafe.Pointer(&root.tree[0]))
	// p := b.ptree
	idx := int16(0)
	for _, c := range []byte(key) {
		k := c & 0x7f
		next := p[idx+int16(k)]
		if next.idx >= 0 {
			tag := root.tags[next.idx]
			if tag.TagName == key {
				return tag
			}
			return nil
		}
		if next.next < 0 {
			return nil
		}
		idx = int16(next.next) * 128
	}

	return nil
}
func (root *tireTree) Get2(key string) *TagInfo {
	status := &root.tree[0]
	for _, c := range []byte(key) {
		k := c & 0x7f
		next := status[k]
		if next.idx >= 0 {
			tag := root.tags[next.idx]
			if tag.TagName == key {
				return tag
			}
			return nil
		}
		if next.next < 0 {
			return nil
		}
		status = &root.tree[next.next]
	}

	return nil
}
