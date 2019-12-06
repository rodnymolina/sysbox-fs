//
// Copyright: (C) 2019 Nestybox Inc.  All rights reserved.
//

package implementations

import (
	"os"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/nestybox/sysbox-fs/domain"
	"github.com/nestybox/sysbox-fs/fuse"
)

//
// /proc/sys/fs/binfmt_misc/register Handler
//
type FsBinfmtRegisterHandler struct {
	Name      string
	Path      string
	Type      domain.HandlerType
	Enabled   bool
	Cacheable bool
	Service   domain.HandlerService
}

func (h *FsBinfmtRegisterHandler) Lookup(n domain.IOnode, pid uint32) (os.FileInfo, error) {

	logrus.Debugf("Executing Lookup() method on %v handler", h.Name)

	return nil, fuse.IOerror{Code: syscall.ENOENT}
}

func (h *FsBinfmtRegisterHandler) Getattr(n domain.IOnode, pid uint32) (*syscall.Stat_t, error) {

	logrus.Debugf("Executing Getattr() method on %v handler", h.Name)

	return nil, nil
}

func (h *FsBinfmtRegisterHandler) Open(n domain.IOnode, pid uint32) error {

	logrus.Debugf("Executing %v Open() method", h.Name)

	return nil
}

func (h *FsBinfmtRegisterHandler) Close(node domain.IOnode) error {

	logrus.Debugf("Executing Close() method on %v handler", h.Name)

	return nil
}

func (h *FsBinfmtRegisterHandler) Read(n domain.IOnode, pid uint32,
	buf []byte, off int64) (int, error) {

	logrus.Debugf("Executing %v Read() method", h.Name)

	return 0, nil
}

func (h *FsBinfmtRegisterHandler) Write(n domain.IOnode, pid uint32,
	buf []byte) (int, error) {

	logrus.Debugf("Executing %v Write() method", h.Name)

	return 0, nil
}

func (h *FsBinfmtRegisterHandler) ReadDirAll(n domain.IOnode, pid uint32) ([]os.FileInfo, error) {

	logrus.Debugf("Executing %v ReadDirAll() method", h.Name)

	return nil, nil
}

func (h *FsBinfmtRegisterHandler) GetName() string {
	return h.Name
}

func (h *FsBinfmtRegisterHandler) GetPath() string {
	return h.Path
}

func (h *FsBinfmtRegisterHandler) GetEnabled() bool {
	return h.Enabled
}

func (h *FsBinfmtRegisterHandler) GetType() domain.HandlerType {
	return h.Type
}

func (h *FsBinfmtRegisterHandler) GetService() domain.HandlerService {
	return h.Service
}

func (h *FsBinfmtRegisterHandler) SetEnabled(val bool) {
	h.Enabled = val
}

func (h *FsBinfmtRegisterHandler) SetService(hs domain.HandlerService) {
	h.Service = hs
}