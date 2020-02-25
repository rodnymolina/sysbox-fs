//
// Copyright: (C) 2019 Nestybox Inc.  All rights reserved.
//

package fuse

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sirupsen/logrus"

	"github.com/nestybox/sysbox-fs/domain"
)

type File struct {
	// File name.
	name string

	// File absolute-path + name.
	path string

	// File attributes.
	attr *fuse.Attr

	// I/O abstraction to represent each file/dir.
	ionode domain.IOnode

	// Pointer to parent fuseService hosting this file/dir.
	service *FuseService
}

//
// NewFile method serves as File constructor.
//
func NewFile(name string, path string, attr *fuse.Attr, srv *FuseService) *File {

	newFile := &File{
		name:    name,
		path:    path,
		attr:    attr,
		service: srv,
		ionode:  srv.ios.NewIOnode(name, path, attr.Mode),
	}

	return newFile
}

//
// Attr FS operation.
//
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {

	logrus.Debug("Requested Attr() operation for entry ", f.path)

	// Simply return the attributes that were previously collected during the
	// lookup() execution.
	*a = *f.attr

	return nil
}

//
// Getattr FS operation.
//
func (f *File) Getattr(
	ctx context.Context,
	req *fuse.GetattrRequest,
	resp *fuse.GetattrResponse) error {

	logrus.Debug("Requested GetAttr() operation for entry ", f.path)

	// Lookup the associated handler within handler-DB.
	handler, ok := f.service.hds.LookupHandler(f.ionode)
	if !ok {
		logrus.Errorf("No supported handler for %v resource", f.path)
		return fmt.Errorf("No supported handler for %v resource", f.path)
	}

	// Handler execution.
	stat, err := handler.Getattr(f.ionode, req.Pid)
	if err != nil {
		logrus.Debug("Getattr() error: ", err)
		return err
	}

	// Simply return the attributes that were previously collected during the
	// lookup() execution, with the exception of the UID/GID, which must be
	// updated based on the obtained response.
	resp.Attr = *f.attr
	if stat != nil {
		resp.Attr.Uid = stat.Uid
		resp.Attr.Gid = stat.Gid
	}

	return nil
}

//
// Open FS operation.
//
func (f *File) Open(
	ctx context.Context,
	req *fuse.OpenRequest,
	resp *fuse.OpenResponse) (fs.Handle, error) {

	logrus.Debug("Requested Open() operation for entry ", f.path)

	f.ionode.SetOpenFlags(int(req.Flags))

	// Lookup the associated handler within handler-DB.
	handler, ok := f.service.hds.LookupHandler(f.ionode)
	if !ok {
		logrus.Errorf("No supported handler for %v resource", f.path)
		return nil, fmt.Errorf("No supported handler for %v resource", f.path)
	}

	// Handler execution.
	err := handler.Open(f.ionode, req.Pid)
	if err != nil && err != io.EOF {
		logrus.Debug("Open() error: ", err)
		return nil, err
	}

	//
	// Due to the nature of procfs and sysfs, files lack explicit sizes (other
	// than zero) as regular files have. In consequence, read operations (also
	// writes) may not be properly handled by kernel, as these ones extend
	// beyond the file sizes reported by Attr() / GetAttr().
	//
	// A solution to this problem is to rely on O_DIRECT flag for all the
	// interactions with procfs/sysfs files. By making use of this flag,
	// sysbox-fs will ensure that it receives all read/write requests
	// generated by fuse-clients, regardless of the file-size issue mentioned
	// above. For regular files, this approach usually comes with a cost, as
	// page-cache is being bypassed for all files I/O; however, this doesn't
	// pose a problem for Inception as we are dealing with special FSs.
	//
	resp.Flags |= fuse.OpenDirectIO

	return f, nil
}

//
// Release FS operation.
//
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {

	logrus.Debug("Requested Release() operation for entry ", f.path)

	// Lookup the associated handler within handler-DB.
	handler, ok := f.service.hds.LookupHandler(f.ionode)
	if !ok {
		logrus.Errorf("No supported handler for %v resource", f.path)
		return fmt.Errorf("No supported handler for %v resource", f.path)
	}

	// Handler execution.
	err := handler.Close(f.ionode)

	return err
}

//
// Read FS operation.
//
func (f *File) Read(
	ctx context.Context,
	req *fuse.ReadRequest,
	resp *fuse.ReadResponse) error {

	logrus.Debug("Requested Read() operation for entry ", f.path)

	if f.ionode == nil {
		logrus.Error("Read() error: File should be properly defined by now")
		return fuse.ENOTSUP
	}

	// Adjust receiving buffer to the request's size.
	resp.Data = resp.Data[:req.Size]

	// Identify the associated handler and execute it accordingly.
	handler, ok := f.service.hds.LookupHandler(f.ionode)
	if !ok {
		logrus.Errorf("Read() error: No supported handler for %v resource", f.path)
		return fmt.Errorf("No supported handler for %v resource", f.path)
	}

	// Handler execution.
	n, err := handler.Read(f.ionode, req.Pid, resp.Data, req.Offset)
	if err != nil && err != io.EOF {
		logrus.Debug("Read() error: ", err)
		return err
	}

	resp.Data = resp.Data[:n]

	return nil
}

//
// Write FS operation.
//
func (f *File) Write(
	ctx context.Context,
	req *fuse.WriteRequest,
	resp *fuse.WriteResponse) error {

	logrus.Debug("Requested Write() operation for entry ", f.path)

	if f.ionode == nil {
		logrus.Error("Write() error: File should be properly defined by now")
		return fuse.ENOTSUP
	}

	// Lookup the associated handler within handler-DB.
	handler, ok := f.service.hds.LookupHandler(f.ionode)
	if !ok {
		logrus.Errorf("Write() error: No supported handler for %v resource", f.path)
		return fmt.Errorf("No supported handler for %v resource", f.path)
	}

	// Handler execution.
	n, err := handler.Write(f.ionode, req.Pid, req.Data)
	if err != nil && err != io.EOF {
		logrus.Debug("Write() error: ", err)
		return err
	}

	resp.Size = n

	return nil
}

//
// Setattr FS operation.
//
func (f *File) Setattr(
	ctx context.Context,
	req *fuse.SetattrRequest,
	resp *fuse.SetattrResponse) error {

	logrus.Debug("Requested Setattr() operation for entry ", f.path)

	// No file attr changes are allowed in a procfs, with the exception of
	// 'size' modifications which are needed to allow write()/truncate() ops.
	// All other 'fuse.SetattrValid' operations will be rejected.
	if req.Valid.Size() {
		return nil
	}

	return fuse.EPERM
}

//
// Forget FS operation.
//
func (f *File) Forget() {

	logrus.Debug("Requested Forget() operation for entry ", f.path)

	f.service.Lock()
	defer f.service.Unlock()

	if _, ok := f.service.nodeDB[f.path]; !ok {
		return
	}

	delete(f.service.nodeDB, f.path)
}

//
// Size method returns the 'size' of a File element.
//
func (f *File) Size() uint64 {
	return f.attr.Size
}

//
// Mode method returns the 'mode' of a File element.
//
func (f *File) Mode() os.FileMode {
	return f.attr.Mode
}

//
// ModTime method returns the modification-time of a File element.
//
func (f *File) ModTime() time.Time {
	return f.attr.Mtime
}

//
// statToAttr helper function to translate FS node-parameters from unix/kernel
// format to FUSE ones.
//
// Kernel FS node attribs:  fuse.attr (fuse_kernel*.go)
// FUSE node attribs:       fuse.Attr (fuse.go)
//
// TODO: Place me in a more appropriate location
//
func statToAttr(s *syscall.Stat_t) fuse.Attr {

	var a fuse.Attr

	a.Inode = uint64(s.Ino)
	a.Size = uint64(s.Size)
	a.Blocks = uint64(s.Blocks)

	a.Atime = time.Unix(s.Atim.Sec, s.Atim.Nsec)
	a.Mtime = time.Unix(s.Mtim.Sec, s.Mtim.Nsec)
	a.Ctime = time.Unix(s.Ctim.Sec, s.Ctim.Nsec)

	a.Mode = os.FileMode(s.Mode)
	a.Nlink = uint32(s.Nlink)
	a.Uid = uint32(s.Uid)
	a.Gid = uint32(s.Gid)
	a.Rdev = uint32(s.Rdev)
	a.BlockSize = uint32(s.Blksize)

	return a
}
