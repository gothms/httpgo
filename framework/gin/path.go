// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE.

package gin

// cleanPath is the URL version of path.Clean, it returns a canonical URL path
// for p, eliminating . and .. elements.
//
// The following rules are applied iteratively until no further processing can
// be done:
//  1. Replace multiple slashes with a single slash.
//  2. Eliminate each . path name element (the current directory).
//  3. Eliminate each inner .. path name element (the parent directory)
//     along with the non-.. element that precedes it.
//  4. Eliminate .. elements that begin a rooted path:
//     that is, replace "/.." by "/" at the beginning of a path.
//
// If the result of this process is an empty string, "/" is returned.
//
// cleanPath 是 path.Clean 的 URL 版本，它为p返回一个规范的URL路径，消除了.和..元素。
//
// 以下规则被反复应用，直到不能再做进一步处理：
// 用一个斜线替换多个斜线。
// 删除每个 . 元素（当前目录）。
// 删除每个内部的 .. 元素（父目录）和它前面的非 .. 元素。
// 删除根路径开始的 .. 元素：也就是说，在一个路径的开始处用 "/" 代替 "/.."。
//
// 如果处理结果是一个空字符串，就会返回 "/"。
func cleanPath(p string) string {
	const stackBufSize = 128 // 默认 128字节
	// Turn empty string into "/"
	if p == "" { // "" 则返回 "/"
		return "/"
	}

	// Reasonably sized buffer on stack to avoid allocations in the common case.
	// If a larger buffer is required, it gets allocated dynamically.
	buf := make([]byte, 0, stackBufSize) // clean 后的结果

	n := len(p)

	// Invariants:
	//      reading from path; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.

	// path must start with '/'
	r := 1 // 下一个要处理的索引
	w := 1 // 结果中药写入的下一个索引

	if p[0] != '/' { // 开头添加 '/'
		r = 0

		if n+1 > stackBufSize {
			buf = make([]byte, n+1)
		} else {
			buf = buf[:n+1]
		}
		buf[0] = '/'
	}

	trailing := n > 1 && p[n-1] == '/' // 是否有 尾斜杠

	// A bit more clunky without a 'lazybuf' like the path package, but the loop
	// gets completely inlined (bufApp calls).
	// loop has no expensive function calls (except 1x make)		// So in contrast to the path package this loop has no expensive function
	// calls (except make, if needed).

	for r < n { // 每次循环，处理一个路径片段（按 / 划分）
		switch {
		case p[r] == '/': // 删除多余的 /
			// empty path element, trailing slash is added after the end
			r++

		case p[r] == '.' && r+1 == n: // . 结尾，通过循环外面的判断来添加尾斜杠
			trailing = true
			r++

		case p[r] == '.' && p[r+1] == '/': // ./ 跳过
			// . element
			r += 2

		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'): // .. 或 ../ 结尾
			// .. element: remove to last /
			r += 3

			if w > 1 { // 回溯到上一个 /
				// can backtrack
				w--

				if len(buf) == 0 {
					for w > 1 && p[w] != '/' { // p[0] 里回溯：p是 / 开头
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' { // clean 过的buf里回溯
						w--
					}
				}
			}

		default:
			// Real path element.
			// Add slash if needed
			if w > 1 { // 路径片段的结尾添加 /
				bufApp(&buf, p, w, '/')
				w++
			}

			// Copy element
			for r < n && p[r] != '/' { // 路径片段写入 buf
				bufApp(&buf, p, w, p[r])
				//fmt.Printf("%d, %c, %v\n", r, p[r], buf)
				w++
				r++
			}
			// s := "abc/ef.g/xy/."：/abc/ef.g/xy/
		}
	}

	// Re-append trailing slash
	if trailing && w > 1 { // 是否添加 尾斜杠
		bufApp(&buf, p, w, '/')
		w++
	}

	// If the original string was not modified (or only shortened at the end),
	// return the respective substring of the original string.
	// Otherwise return a new string from the buffer.
	if len(buf) == 0 {
		return p[:w]
	}
	return string(buf[:w])
}

// Internal helper to lazily create a buffer if necessary.
// Calls to this function get inlined.
// 参数：缓存 原字符串 s索引/写入位置 要写入的字符
func bufApp(buf *[]byte, s string, w int, c byte) {
	b := *buf
	// 循环是完全内联的：https://juejin.cn/post/7128421990943162404
	// $ go build -gcflags="-m -m" .\path.go
	// 这里实际上 go 的编译器优化，对于在 AST 中节点数小于 80 个的函数会自动进行内联优化
	// 所以 bufApp 就被自动内联到循环中了
	if len(b) == 0 {
		// No modification of the original string so far.
		// If the next character is the same as in the original string, we do
		// not yet have to allocate a buffer.
		if s[w] == c { // 写入字符与原字符串在该位置字符相同，则直接返回
			return // 路径需要被化简时，才会创建缓存
		}

		// Otherwise use either the stack buffer, if it is large enough, or
		// allocate a new buffer on the heap, and copy all previous characters.
		length := len(s)
		if length > cap(b) { // 容量不足
			*buf = make([]byte, length) // 创建 buf
		} else {
			*buf = (*buf)[:length]
		}
		b = *buf

		copy(b, s[:w]) // 拷贝前面（首次创建）
	}
	b[w] = c // 写入字符
}
