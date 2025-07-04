// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package inmemory

import (
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"
)

var (
	errExists    = fmt.Errorf("exists")
	errNotExists = fmt.Errorf("notexists")
	errIsNotDir  = fmt.Errorf("notdir")
	errIsDir     = fmt.Errorf("isdir")
)

type node interface {
	name() string
	path() string
	isdir() bool
	modtime() time.Time
}

// dir is the central type for the memory-based  storagedriver. All operations
// are dispatched from a root dir.
type dir struct {
	common

	// TODO(stevvooe): Use sorted slice + search.
	children map[string]node
}

var _ node = &dir{}

func (d *dir) isdir() bool {
	return true
}

// add places the node n into dir d.
func (d *dir) add(n node) {
	if d.children == nil {
		d.children = make(map[string]node)
	}

	d.children[n.name()] = n
	d.mod = time.Now()
}

// find searches for the node, given path q in dir. If the node is found, it
// will be returned. If the node is not found, the closet existing parent. If
// the node is found, the returned (node).path() will match q.
func (d *dir) find(q string) node {
	q = strings.Trim(q, "/")
	i := strings.Index(q, "/")

	if q == "" {
		return d
	}

	if i == 0 {
		panic("shouldn't happen, no root paths")
	}

	var component string
	if i < 0 {
		// No more path components
		component = q
	} else {
		component = q[:i]
	}

	child, ok := d.children[component]
	if !ok {
		// Node was not found. Return p and the current node.
		return d
	}

	if child.isdir() {
		// traverse down!
		q = q[i+1:]
		return child.(*dir).find(q)
	}

	return child
}

func (d *dir) list(p string) ([]string, error) {
	n := d.find(p)

	if n.path() != p {
		return nil, errNotExists
	}

	if !n.isdir() {
		return nil, errIsNotDir
	}

	// NOTE(milosgajdos): this is safe to do because
	// n can only be *dir due to the compile time check
	dirChildren := n.(*dir).children

	children := make([]string, 0, len(dirChildren))
	for _, child := range dirChildren {
		children = append(children, child.path())
	}

	sort.Strings(children)
	return children, nil
}

// mkfile or return the existing one. returns an error if it exists and is a
// directory. Essentially, this is open or create.
func (d *dir) mkfile(p string) (*file, error) {
	n := d.find(p)
	if n.path() == p {
		if n.isdir() {
			return nil, errIsDir
		}

		return n.(*file), nil
	}

	dirpath, filename := path.Split(p)
	// Make any non-existent directories
	n, err := d.mkdirs(dirpath)
	if err != nil {
		return nil, err
	}

	dd := n.(*dir)
	n = &file{
		common: common{
			p:   path.Join(dd.path(), filename),
			mod: time.Now(),
		},
	}

	dd.add(n)
	return n.(*file), nil
}

// mkdirs creates any missing directory entries in p and returns the result.
func (d *dir) mkdirs(p string) (*dir, error) {
	p = normalize(p)

	n := d.find(p)

	if !n.isdir() {
		// Found something there
		return nil, errIsNotDir
	}

	if n.path() == p {
		return n.(*dir), nil
	}

	dd := n.(*dir)

	relative := strings.Trim(strings.TrimPrefix(p, n.path()), "/")

	if relative == "" {
		return dd, nil
	}

	components := strings.Split(relative, "/")
	for _, component := range components {
		d, err := dd.mkdir(component)
		if err != nil {
			// This should actually never happen, since there are no children.
			return nil, err
		}
		dd = d
	}

	return dd, nil
}

// mkdir creates a child directory under d with the given name.
func (d *dir) mkdir(name string) (*dir, error) {
	if name == "" {
		return nil, fmt.Errorf("invalid dirname")
	}

	_, ok := d.children[name]
	if ok {
		return nil, errExists
	}

	child := &dir{
		common: common{
			p:   path.Join(d.path(), name),
			mod: time.Now(),
		},
	}
	d.add(child)
	d.mod = time.Now()

	return child, nil
}

func (d *dir) move(src, dst string) error {
	dstDirname, _ := path.Split(dst)

	dp, err := d.mkdirs(dstDirname)
	if err != nil {
		return err
	}

	srcDirname, srcFilename := path.Split(src)
	sp := d.find(srcDirname)

	if normalize(srcDirname) != normalize(sp.path()) {
		return errNotExists
	}

	spd, ok := sp.(*dir)
	if !ok {
		return errIsNotDir // paranoid.
	}

	s, ok := spd.children[srcFilename]
	if !ok {
		return errNotExists
	}

	delete(spd.children, srcFilename)

	switch n := s.(type) {
	case *dir:
		n.p = dst
	case *file:
		n.p = dst
	}

	dp.add(s)

	return nil
}

func (d *dir) delete(p string) error {
	dirname, filename := path.Split(p)
	parent := d.find(dirname)

	if normalize(dirname) != normalize(parent.path()) {
		return errNotExists
	}

	parentDir, ok := parent.(*dir)
	if !ok {
		return errIsNotDir
	}

	if _, ok := parentDir.children[filename]; !ok {
		return errNotExists
	}

	delete(parentDir.children, filename)
	return nil
}

func (d *dir) String() string {
	return fmt.Sprintf("&dir{path: %v, children: %v}", d.p, d.children)
}

// file stores actual data in the fs tree. It acts like an open, seekable file
// where operations are conducted through ReadAt and WriteAt. Use it with
// SectionReader for the best effect.
type file struct {
	common
	data []byte
}

var _ node = &file{}

func (f *file) isdir() bool {
	return false
}

func (f *file) truncate() {
	f.data = f.data[:0]
}

func (f *file) sectionReader(offset int64) io.Reader {
	return io.NewSectionReader(f, offset, int64(len(f.data))-offset)
}

func (f *file) ReadAt(p []byte, offset int64) (n int, err error) {
	if offset >= int64(len(f.data)) {
		return 0, io.EOF
	}
	return copy(p, f.data[offset:]), nil
}

// reallocExponent is the exponent used to realloc a slice. The value roughly
// follows the behavior of Go built-in append function.
const reallocExponent = 1.25

func (f *file) WriteAt(p []byte, offset int64) (n int, err error) {
	newLen := offset + int64(len(p))
	if int64(cap(f.data)) < newLen {
		// Grow slice exponentially to ensure amortized linear time complexity
		// of reallocation
		newCap := int64(float64(cap(f.data)) * reallocExponent)
		if newCap < newLen {
			newCap = newLen
		}
		data := make([]byte, len(f.data), newCap)
		copy(data, f.data)
		f.data = data
	}

	f.mod = time.Now()
	f.data = f.data[:newLen]

	return copy(f.data[offset:newLen], p), nil
}

func (f *file) String() string {
	return fmt.Sprintf("&file{path: %q}", f.p)
}

// common provides shared fields and methods for node implementations.
type common struct {
	p   string
	mod time.Time
}

func (c *common) name() string {
	_, name := path.Split(c.p)
	return name
}

func (c *common) path() string {
	return c.p
}

func (c *common) modtime() time.Time {
	return c.mod
}

func normalize(p string) string {
	return "/" + strings.Trim(p, "/")
}
