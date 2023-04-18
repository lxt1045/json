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
)

type binTreeNode struct {
	next uint16 // 下一个状态
	// idx  int16  // 只有 '"' 才是借宿标志，才有 idx
	tag *TagInfo
}
type binTree struct {
	tree [][128]binTreeNode // 状态
	// tags []*TagInfo
}

func NewBinTree(tags []*TagInfo) (root *binTree, err error) {
	for i, tag := range tags {
		tags[i].TagName = tag.TagName + `"`
	}
	root = &binTree{
		tree: make([][128]binTreeNode, 1, 4),
		// tags: tags,
	}

out:
	for _, tag := range tags {
		key := tag.TagName
		status := &root.tree[0]
		for iKey, c := range []byte(key) {
			k := c % 128
			n := &status[k]

			// 没有被占领或是叶子结点
			if n.next == 0 {
				// 没有被占领
				if n.tag == nil {
					//占领此叶节点
					n.tag = tag
					continue out
				}
				// 叶子节点
				old := n.tag.TagName
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
				n.tag = nil
				n.next = uint16(len(root.tree))

				// 给旧的 node 添加状态
				root.tree = append(root.tree, [128]binTreeNode{})
				status = &root.tree[len(root.tree)-1] // 创建新的状态

				kOld := old[iKey+1] % 128
				status[kOld] = nOld

				// key 的 next 在 for 的下一轮再处理！
				continue
			}
			status = &root.tree[n.next]
		}
	}

	//统一处理: idx --
	return
}

func (b *binTree) Get(key string) *TagInfo {
	status := &b.tree[0]
	for _, c := range []byte(key) {
		k := c & 0x7f
		next := status[k]
		if next.tag != nil {
			if next.tag.TagName == key {
				return next.tag
			}
			return nil
		}
		if next.next == 0 {
			return nil
		}
		status = &b.tree[next.next]
	}

	return nil
}
