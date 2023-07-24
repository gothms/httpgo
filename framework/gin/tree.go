// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE

package gin

import (
	"bytes"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gothms/httpgo/framework/gin/internal/bytesconv"
)

var (
	strColon = []byte(":")
	strStar  = []byte("*")
	strSlash = []byte("/")
)

// Param is a single URL parameter, consisting of a key and a value.
// 存储路由参数
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// Get returns the value of the first Param which key matches the given name and a boolean true.
// If no matching Param is found, an empty string is returned and a boolean false .
func (ps Params) Get(name string) (string, bool) {
	for _, entry := range ps {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return "", false
}

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) (va string) {
	va, _ = ps.Get(name)
	return
}

type methodTree struct {
	method string // http method
	root   *node  // 根节点
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// addChild will add a child node, keeping wildcardChild at the end
// addChild将添加一个子节点，最后保留通配符Child。
func (n *node) addChild(child *node) {
	if n.wildChild && len(n.children) > 0 {
		wildcardChild := n.children[len(n.children)-1]
		n.children = append(n.children[:len(n.children)-1], child, wildcardChild)
	} else {
		n.children = append(n.children, child)
	}
}

func countParams(path string) uint16 {
	var n uint16
	s := bytesconv.StringToBytes(path)
	n += uint16(bytes.Count(s, strColon)) // ':' bytealg.Count(s, sep[0])
	n += uint16(bytes.Count(s, strStar))  // '*'
	return n
}

func countSections(path string) uint16 {
	s := bytesconv.StringToBytes(path)
	return uint16(bytes.Count(s, strSlash)) // '/'
}

type nodeType uint8

const (
	static   nodeType = iota // 默认，普通节点
	root                     // 根
	param                    // 通配符节点：: 命名参数捕获
	catchAll                 // 通配符节点：* 任意参数捕获
)

type node struct {
	path      string        // 节点path，即路径片段
	indices   string        // 子节点“索引”
	wildChild bool          // 子节点是否为通配符节点 : *
	nType     nodeType      // 节点类型
	priority  uint32        // 优先级，子节点数量越多，优先级越高
	children  []*node       // child nodes, at most 1 :param style node at the end of the array
	handlers  HandlersChain // 节点路径的handle
	fullPath  string        // 全路径，可能为 ""，仅用于报错
}

// Increments priority of the given child and reorders if necessary
// 给定子节点的优先级，必要时重新排序
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++ // 优先级增加
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- { // 插入排序 & 快排
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Build new index char string
	if newPos != pos { // 根据priority重新排序 “索引”：4段字符串 + 低效
		n.indices = n.indices[:newPos] + // Unchanged prefix, might be empty
			n.indices[pos:pos+1] + // The index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // Rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given handle to the path.
// Not concurrency-safe!
// addRoute向路径中添加一个具有给定句柄的节点。
// 不是并发安全的!
func (n *node) addRoute(path string, handlers HandlersChain) {
	fullPath := path
	n.priority++

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, fullPath, handlers)
		n.nType = root
		return
	}

	parentFullPathIndex := 0

walk:
	for {
		// Find the longest common prefix.
		// This also implies that the common prefix contains no ':' or '*'
		// since the existing key can't contain those chars.
		i := longestCommonPrefix(path, n.path) // 公共前缀

		// Split edge
		if i < len(n.path) { // 新公共前缀长度 < 原公共前缀长度，原node一分为二
			child := node{
				path:      n.path[i:], // 子节点：原节点path的后半部分
				wildChild: n.wildChild,
				nType:     static,
				indices:   n.indices,
				children:  n.children, // 原子节点们
				handlers:  n.handlers,
				priority:  n.priority - 1,
				fullPath:  n.fullPath,
			}

			n.children = []*node{&child} // 因为 child 仍然是所有以child为前缀的路径的公共前缀
			// []byte for proper unicode char conversion, see #65
			n.indices = bytesconv.BytesToString([]byte{n.path[i]})
			n.path = path[:i] // 当前节点：原节点path的前半部分
			n.handlers = nil
			n.wildChild = false
			n.fullPath = fullPath[:parentFullPathIndex+i]
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:] // 除去公共前缀
			c := path[0]

			// '/' after param
			if n.nType == param && c == '/' && len(n.children) == 1 {
				parentFullPathIndex += len(n.path)
				n = n.children[0] // 只有一个子节点
				n.priority++
				continue walk
			} // 参数节点：n是':'后面的参数，后面接 '/'

			// Check if a child with the next path byte exists
			for i, max := 0, len(n.indices); i < max; i++ {
				if c == n.indices[i] {
					parentFullPathIndex += len(n.path)
					i = n.incrementChildPrio(i)
					n = n.children[i] // 和path有公共前缀的子节点
					continue walk
				} // 和子节点有公共前缀
			}

			// Otherwise insert it：否则直接插入
			if c != ':' && c != '*' && n.nType != catchAll {
				// []byte for proper unicode char conversion, see #65
				n.indices += bytesconv.BytesToString([]byte{c}) // 添加索引
				child := &node{
					fullPath: fullPath,
				}
				n.addChild(child)
				n.incrementChildPrio(len(n.indices) - 1) // n.priority++：并重新索引排序
				n = child
			} else if n.wildChild { // 同为参数节点 : 或 *
				// inserting a wildcard node, need to check if it conflicts with the existing wildcard
				n = n.children[len(n.children)-1] // 参数节点为最后一个子节点
				n.priority++

				// Check if the wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					// Adding a child to a catchAll is not possible：catchAll 节点不能有子节点
					n.nType != catchAll &&
					// Check for longer wildcard, e.g. :name and :names
					//(len(n.path) >= len(path) || path[len(n.path)] == '/') { // 逻辑判断不严谨 TODO
					(len(n.path) == len(path) || path[len(n.path)] == '/') {
					continue walk // 继续查找，并去掉公共前缀
				} // 参数节点完全相同

				// Wildcard conflict：冲突
				pathSeg := path          // 新路径出现冲突的部分（没有完成插入的部分）
				if n.nType != catchAll { // 从冲突部分的开始到下一个'/'之前，为错误路径
					pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
				}
				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path // 树中包含冲突节点的前缀路径
				panic("'" + pathSeg +
					"' in new path '" + fullPath +
					"' conflicts with existing wildcard '" + n.path +
					"' in existing prefix '" + prefix +
					"'") // 参数节点不同
			}

			n.insertChild(path, fullPath, handlers) // 直接插入新节点
			return
		}

		// Otherwise add handle to current node
		if n.handlers != nil { // 该path已被注册过
			panic("handlers are already registered for path '" + fullPath + "'")
		}
		n.handlers = handlers
		n.fullPath = fullPath
		return
	}
}

// Search for a wildcard segment and check the name for invalid characters.
// Returns -1 as index, if no wildcard was found.
// 搜索通配符段并检查名称中是否有无效字符。
// 如果没有找到通配符，则返回-1作为索引。
func findWildcard(path string) (wildcard string, i int, valid bool) {
	// Find start
	//fmt.Println("findWildcard", path)
	for start, c := range []byte(path) {
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

// insertChild addRoute 和 insertChild 两个功能解耦得非常干净
// addRoute 函数本身的代码只负责对公共前缀的查找和对路由中已有路径的节点进行修改，不会涉及到新路径节点的添加
func (n *node) insertChild(path string, fullPath string, handlers HandlersChain) {
	for {
		// Find prefix until first wildcard
		wildcard, i, valid := findWildcard(path)
		//fmt.Println(wildcard, i, valid, path)

		if i < 0 { // No wildcard found
			break
		}

		// The wildcard name must only contain one ':' or '*' character
		// 大于 1 个 :/*
		if !valid { // 每个路径段只允许一个通配符
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// check if the wildcard has a name
		if len(wildcard) < 2 { // :/* 后面是"空"，没有值
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if wildcard[0] == ':' { // param
			if i > 0 { // i == 0：:前面没有路径
				// Insert prefix before the current wildcard
				// 在当前通配符前插入前缀
				n.path = path[:i]
				path = path[i:]
			}

			child := &node{
				nType:    param,
				path:     wildcard,
				fullPath: fullPath,
			}
			n.addChild(child)
			n.wildChild = true // : 前的节点
			n = child
			n.priority++

			// if the path doesn't end with the wildcard, then there
			// will be another subpath starting with '/'
			// 如果路径没有以通配符结尾，那么就会出现 将有另一个以'/'开头的子路径
			if len(wildcard) < len(path) {
				//fmt.Println(wildcard, path)
				path = path[len(wildcard):]

				child := &node{
					priority: 1,
					fullPath: fullPath,
				}
				n.addChild(child)
				n = child
				//fmt.Println("path:", n.path)
				continue
			}

			// Otherwise we're done. Insert the handle in the new leaf
			n.handlers = handlers
			return
		}

		// catchAll：通配符 *
		if i+len(wildcard) != len(path) { // 在路径中，只允许在路径的末端有 * 路由。
			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
		}

		// 参数节点不能与其他节点共存于路径的同一位置中，会发生冲突
		// 例如："/who/are/you" 和 "/who/are/*x"
		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' { // 节点结尾是 '/'
			//fmt.Printf("error: %s, %c", n.path, n.path[len(n.path)-1])
			pathSeg := strings.SplitN(n.children[0].path, "/", 2)[0]
			panic("catch-all wildcard '" + path +
				"' in new path '" + fullPath +
				"' conflicts with existing path segment '" + pathSeg +
				"' in existing prefix '" + n.path + pathSeg +
				"'")
		}

		// currently fixed width 1 for '/'
		i--
		if path[i] != '/' { // * 前需要有 /，错误示例 /a*a
			panic("no / before catch-all in path '" + fullPath + "'")
		}

		// 从第一个 / 切开，分两段添加到 tree
		n.path = path[:i]

		// First node: catchAll node with empty path
		child := &node{ // "" 字符串
			wildChild: true, // * 前一个节点：true
			nType:     catchAll,
			fullPath:  fullPath,
		}

		n.addChild(child)
		n.indices = string('/') // 索引
		n = child
		n.priority++

		// second node: node holding the variable
		child = &node{ // 通配符path：头两个字符肯定是 /*
			path:     path[i:],
			nType:    catchAll,
			handlers: handlers,
			priority: 1,
			fullPath: fullPath,
		}
		n.children = []*node{child}

		return
	}

	// If no wildcard was found, simply insert the path and handle
	n.path = path // 可能是 /
	n.handlers = handlers
	n.fullPath = fullPath
}

// nodeValue holds return values of (*Node).getValue method
// 路由匹配返回的结构体
type nodeValue struct {
	handlers HandlersChain // 处理函数集
	params   *Params       // 请求承诺书
	tsr      bool          // 尾部斜线/尾斜杠重定向：如果查找的路径增加或删除尾斜杠后存在 handle，标记 tsr 为 true
	fullPath string        // 请求的完整路径
}

type skippedNode struct {
	path        string
	node        *node
	paramsCount int16
}

// Returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
// 返回以给定路径（key）注册的句柄。通配符的值被保存到一个map中。
// 如果找不到句柄，如果有一个句柄存在于给定路径的额外的（没有）尾部斜杠，将进行TSR（尾部斜杠重定向）推荐。
func (n *node) getValue(path string, params *Params, skippedNodes *[]skippedNode, unescape bool) (value nodeValue) {
	var globalParamsCount int16

walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if len(path) > len(prefix) { // path长度 > 当前节点路径长度
			if path[:len(prefix)] == prefix { // 当前节点匹配
				path = path[len(prefix):] // 待匹配路径

				// Try all the non-wildcard children first by matching the indices
				idxc := path[0]
				for i, c := range []byte(n.indices) { // 遍历子节点的索引
					if c == idxc { // 索引匹配
						//  strings.HasPrefix(n.children[len(n.children)-1].path, ":") == n.wildChild
						if n.wildChild {
							index := len(*skippedNodes)
							*skippedNodes = (*skippedNodes)[:index+1]
							(*skippedNodes)[index] = skippedNode{
								path: prefix + path,
								node: &node{
									path:      n.path,
									wildChild: n.wildChild,
									nType:     n.nType,
									priority:  n.priority,
									children:  n.children,
									handlers:  n.handlers,
									fullPath:  n.fullPath,
								},
								paramsCount: globalParamsCount,
							}
						}

						n = n.children[i] // 从这个索引开始匹配
						continue walk
					}
				}

				if !n.wildChild { // 当前节点不是参数节点
					// If the path at the end of the loop is not equal to '/' and the current node has no child nodes
					// the current node needs to roll back to last valid skippedNode
					if path != "/" {
						for length := len(*skippedNodes); length > 0; length-- {
							skippedNode := (*skippedNodes)[length-1]
							*skippedNodes = (*skippedNodes)[:length-1]
							if strings.HasSuffix(skippedNode.path, path) {
								path = skippedNode.path
								n = skippedNode.node
								if value.params != nil {
									*value.params = (*value.params)[:skippedNode.paramsCount]
								}
								globalParamsCount = skippedNode.paramsCount
								continue walk
							}
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					value.tsr = path == "/" && n.handlers != nil // tsr 的条件：尾斜线 + 存在 handlers
					return
				}

				// Handle wildcard child, which is always at the end of the array
				n = n.children[len(n.children)-1] // : 和 * 通配符匹配
				globalParamsCount++

				switch n.nType {
				case param: // 参数匹配
					// fix truncate the parameter
					// tree_test.go  line: 204

					// Find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// Save param value
					if params != nil && cap(*params) > 0 {
						if value.params == nil {
							value.params = params
						}
						// Expand slice within preallocated capacity
						i := len(*value.params)
						*value.params = (*value.params)[:i+1] // 扩容。params 的值的设置：func (engine *Engine) addRoute
						val := path[:end]
						if unescape {
							if v, err := url.QueryUnescape(val); err == nil {
								val = v
							}
						}
						(*value.params)[i] = Param{
							Key:   n.path[1:], // 去掉 ':'
							Value: val,
						}
					}

					// we need to go deeper!
					if end < len(path) { // 路径未匹配完成
						if len(n.children) > 0 { // 可以继续匹配
							path = path[end:]
							n = n.children[0] // 第一个子节点
							continue walk
						}

						// ... but we can't：没有子节点了
						value.tsr = len(path) == end+1 // 路径以 '/' 结尾，tsr = true
						return
					}
					// end == len(path)
					if value.handlers = n.handlers; value.handlers != nil {
						value.fullPath = n.fullPath
						return
					}
					if len(n.children) == 1 { // n.handlers == nil
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						// 子节点为 "/" && handlers != nil 或者 * 匹配：tsr=true
						value.tsr = (n.path == "/" && n.handlers != nil) || (n.path == "" && n.indices == "/")
					}
					return

				case catchAll:
					// Save param value
					if params != nil {
						if value.params == nil {
							value.params = params
						}
						// Expand slice within preallocated capacity
						i := len(*value.params)
						*value.params = (*value.params)[:i+1]
						val := path
						if unescape {
							if v, err := url.QueryUnescape(path); err == nil {
								val = v
							}
						}
						(*value.params)[i] = Param{
							Key:   n.path[2:], // 去掉 /*
							Value: val,
						}
					}

					value.handlers = n.handlers
					value.fullPath = n.fullPath
					return

				default:
					panic("invalid node type")
				}
			}
		}

		if path == prefix { // 路径匹配成功
			// If the current path does not equal '/' and the node does not have a registered handle and the most recently matched node has a child node
			// the current node needs to roll back to last valid skippedNode
			if n.handlers == nil && path != "/" {
				for length := len(*skippedNodes); length > 0; length-- {
					skippedNode := (*skippedNodes)[length-1]
					*skippedNodes = (*skippedNodes)[:length-1]
					if strings.HasSuffix(skippedNode.path, path) {
						path = skippedNode.path
						n = skippedNode.node
						if value.params != nil {
							*value.params = (*value.params)[:skippedNode.paramsCount]
						}
						globalParamsCount = skippedNode.paramsCount
						continue walk
					}
				}
				//	n = latestNode.children[len(latestNode.children)-1]
			}
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if value.handlers = n.handlers; value.handlers != nil {
				value.fullPath = n.fullPath
				return
			} // handlers 存在

			// If there is no handle for this route, but this route has a
			// wildcard child, there must be a handle for this path with an
			// additional trailing slash
			if path == "/" && n.wildChild && n.nType != root {
				value.tsr = true
				return
			} // 注册了 "/con/:tact"，匹配 "/con/"：tsr=true 并去掉 '/'

			if path == "/" && n.nType == static {
				value.tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, c := range []byte(n.indices) {
				if c == '/' { // '/' 或 /* 通配符：tsr=true，并加上 '/'
					n = n.children[i]
					value.tsr = (len(n.path) == 1 && n.handlers != nil) ||
						(n.nType == catchAll && n.children[0].handlers != nil)
					return
				}
			}

			return
		}
		// 匹配失败。path=="/" 或者 当前路径+'/' 即可匹配上一个带 handlers 的路由时，value.tsr = true
		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		value.tsr = path == "/" ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.handlers != nil)

		// roll back to last valid skippedNode
		if !value.tsr && path != "/" {
			for length := len(*skippedNodes); length > 0; length-- {
				skippedNode := (*skippedNodes)[length-1]
				*skippedNodes = (*skippedNodes)[:length-1]
				if strings.HasSuffix(skippedNode.path, path) {
					path = skippedNode.path
					n = skippedNode.node
					if value.params != nil {
						*value.params = (*value.params)[:skippedNode.paramsCount]
					}
					globalParamsCount = skippedNode.paramsCount
					continue walk
				}
			} // 回滚到最后一个有效的skippedNode
		}

		return
	}
}

// Makes a case-insensitive lookup of the given path and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) ([]byte, bool) {
	const stackBufSize = 128 // 默认 128 个字节

	// Use a static sized buffer on the stack in the common case.
	// If the path is too long, allocate a buffer on the heap instead.
	buf := make([]byte, 0, stackBufSize)
	if length := len(path) + 1; length > stackBufSize {
		buf = make([]byte, 0, length) // 为新路径分配足够的空间
	}

	ciPath := n.findCaseInsensitivePathRec( // 大小写非敏感查找
		path,
		buf,              // Preallocate enough memory for new path
		[4]byte{},        // Empty rune buffer：空的 rune 缓存
		fixTrailingSlash, // 可选配置，由路由中是否启用了 RedirectTrailingSlash 功能决定，它和 RedirectFixedPath 功能是独立的
	)

	return ciPath, ciPath != nil
}

// Shift bytes in array by n bytes left
func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	switch n {
	// 左移 n “位”：因为 utf-8 字符在 string 中按字节存储，它的长度从 1 字节到 4 字节不等
	// 为了处理 utf-8 字符，需要这样一个对 4 字节的缓存的操作方法
	case 0:
		return rb
	case 1:
		return [4]byte{rb[1], rb[2], rb[3], 0}
	case 2:
		return [4]byte{rb[2], rb[3]}
	case 3:
		return [4]byte{rb[3]}
	default: // n>3
		return [4]byte{}
	}
}

// Recursive case-insensitive lookup function used by n.findCaseInsensitivePath
// n.findCaseInsensitivePath 使用的不区分大小写的路由查找。Rec 意思是 Recursive
func (n *node) findCaseInsensitivePathRec(path string, ciPath []byte, rb [4]byte, fixTrailingSlash bool) []byte {
	npLen := len(n.path)
	// EqualFold 对两个字符串进行大小写非敏感的匹配

walk: // Outer loop for walking the tree
	// 未匹配路径 >= 节点路径 && (通配符 || 节点路径匹配)：从 1 开始，是因为 0 是索引
	for len(path) >= npLen && (npLen == 0 || strings.EqualFold(path[1:npLen], n.path[1:])) {
		// Add common prefix to result
		oldPath := path // 记录 path
		path = path[npLen:]
		ciPath = append(ciPath, n.path...) // 修正路径，最终返回的路径

		if len(path) == 0 {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.handlers != nil {
				return ciPath
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash { // Engine.RedirectFixedPath：入参时为 true
				for i, c := range []byte(n.indices) {
					if c == '/' { // 有 '/' 索引
						n = n.children[i]
						if (len(n.path) == 1 && n.handlers != nil) ||
							(n.nType == catchAll && n.children[0].handlers != nil) {
							return append(ciPath, '/')
						} // 子节点为：'/' 或 /* 通配符
						return nil
					}
				}
			}
			return nil
		}
		/*
			实际上路径中是可以出现如中文之类的字符的
			所以路由树在构建过程中划分时是按照字节进行划分，而不是按照字符进行划分
			也就是说在一个节点中可能存储的不是完整的字符路径
		*/

		// If this node does not have a wildcard (param or catchAll) child,
		// we can just look up the next child node and continue to walk down
		// the tree
		if !n.wildChild { // 非参数节点
			// Skip rune bytes already processed
			rb = shiftNRuneBytes(rb, npLen)

			if rb[0] != 0 { // 缓存中还有字节未处理：一个utf8字符被分割为多个节点
				// Old rune not finished
				idxc := rb[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						// continue with child node
						n = n.children[i]
						npLen = len(n.path)
						continue walk
					}
				}
			} else { // 缓存中字节已处理完成（缓存为空）
				// Process a new rune
				var rv rune

				// Find rune start.
				// Runes are up to 4 byte long,
				// -4 would definitely be another rune.
				var off int
				for max := min(npLen, 3); off < max; off++ { // 处理utf8，比如中文
					// utf8.RuneStart：判断传入字节是否为某个字符的开始
					// utf8.DecodeRuneInString：从字符串中解码出第一个字符
					if i := npLen - off; utf8.RuneStart(oldPath[i]) {
						// read rune from cached path
						rv, _ = utf8.DecodeRuneInString(oldPath[i:])
						break
					}
				}

				// Calculate lowercase bytes of current rune
				lo := unicode.ToLower(rv)  // 小写匹配
				utf8.EncodeRune(rb[:], lo) // 字符转字节

				// Skip already processed bytes
				rb = shiftNRuneBytes(rb, off) // 跳过已处理过的字节

				idxc := rb[0]
				for i, c := range []byte(n.indices) {
					// Lowercase matches
					if c == idxc { // 匹配到索引
						// must use a recursive approach since both the
						// uppercase byte and the lowercase byte might exist
						// as an index
						if out := n.children[i].findCaseInsensitivePathRec( // 子节点
							path, ciPath, rb, fixTrailingSlash,
						); out != nil {
							return out
						} // 递归分支：大写/小写，分支匹配失败，则回溯
						break
					}
				}

				// If we found no match, the same for the uppercase rune,
				// if it differs
				if up := unicode.ToUpper(rv); up != lo { // 大写匹配
					utf8.EncodeRune(rb[:], up)    // 字符转字节
					rb = shiftNRuneBytes(rb, off) // 跳过已处理过的字节

					idxc := rb[0]
					for i, c := range []byte(n.indices) {
						// Uppercase matches
						if c == idxc {
							// Continue with child node
							n = n.children[i]
							npLen = len(n.path) // 继续循环
							continue walk
						} // 不用递归了
					}
				}
			}

			// Nothing found. We can recommend to redirect to the same URL
			// without a trailing slash if a leaf exists for that path
			if fixTrailingSlash && path == "/" && n.handlers != nil {
				return ciPath // 尾部斜线
			}
			return nil
		}

		// : 或 *
		n = n.children[0]
		switch n.nType {
		case param: // 命名捕获参数
			// 相对 getValue 中的逻辑几乎一样
			// 只是删除了存储参数的代码
			// 并增加了记录路径的代码（因为这部分路径是参数，所以不需要大小写修正）
			// Find param end (either '/' or path end)
			end := 0
			for end < len(path) && path[end] != '/' {
				end++
			}

			// Add param value to case insensitive path
			ciPath = append(ciPath, path[:end]...) // 记录路径

			// We need to go deeper!
			if end < len(path) { // 路径未匹配完成
				if len(n.children) > 0 { // 可以继续匹配
					// Continue with child node
					n = n.children[0]   // 第一个子节点
					npLen = len(n.path) // npLen赋值
					path = path[end:]
					continue
				}

				// ... but we can't：没有子节点了
				if fixTrailingSlash && len(path) == end+1 { // 路径以 '/' 结尾，tsr = true
					return ciPath // 返回修正后的路径
				}
				return nil
			}

			if n.handlers != nil {
				return ciPath
			}

			if fixTrailingSlash && len(n.children) == 1 {
				// No handle found. Check if a handle for this path + a
				// trailing slash exists
				n = n.children[0]
				if n.path == "/" && n.handlers != nil {
					return append(ciPath, '/')
				} // 子节点为 "/"
			}

			return nil

		case catchAll: // 通配符
			return append(ciPath, path...)

		default:
			panic("invalid node type")
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash { // 匹配失败，尝试添加/删除 尾部斜线
		if path == "/" {
			return ciPath
		}
		// 路径和子节点匹配，但是子节点多一个 '/'，且 handlers 存在
		if len(path)+1 == npLen && n.path[len(path)] == '/' &&
			strings.EqualFold(path[1:], n.path[1:len(path)]) && n.handlers != nil {
			return append(ciPath, n.path...)
		}
	}
	return nil
}
