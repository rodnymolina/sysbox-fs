//
// Copyright: (C) 2019 Nestybox Inc.  All rights reserved.
//

package implementations

import (
	"errors"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/nestybox/sysbox-fs/domain"
)

//
// Due to the fact that sysbox-fs' procfs is sourced at /proc/sys, there's no
// much this handler needs to do. This handler's purpose is to be able to manage
// operations associated to /proc bind-mounts such as cpuinfo, meminfo, etc).
//

//
// /proc Handler
//
type SysCommonHandler struct {
	Name      string
	Path      string
	Type      domain.HandlerType
	Enabled   bool
	Cacheable bool
	Service   domain.HandlerService
}

func (h *SysCommonHandler) Lookup(n domain.IOnode, pid uint32) (os.FileInfo, error) {

	logrus.Debugf("Executing Lookup() method on %v handler", h.Name)

	// Identify the pidNsInode corresponding to this pid.
	pidInode := h.Service.FindPidNsInode(pid)
	if pidInode == 0 {
		return nil, errors.New("Could not identify pidNsInode")
	}

	return n.Stat()
}

func (h *SysCommonHandler) Getattr(n domain.IOnode, pid uint32) (*syscall.Stat_t, error) {

	logrus.Debugf("Executing Getattr() method on %v handler", h.Name)

	return nil, nil
}

func (h *SysCommonHandler) Open(n domain.IOnode, pid uint32) error {

	logrus.Debugf("Executing %v Open() method", h.Name)

	return nil
}

func (h *SysCommonHandler) Close(n domain.IOnode) error {

	logrus.Debugf("Executing Close() method on %v handler", h.Name)

	return nil
}

func (h *SysCommonHandler) Read(n domain.IOnode, pid uint32,
	buf []byte, off int64) (int, error) {

	logrus.Debugf("Executing %v Read() method", h.Name)

	return 0, nil
}

func (h *SysCommonHandler) Write(n domain.IOnode, pid uint32,
	buf []byte) (int, error) {

	logrus.Debugf("Executing %v Write() method", h.Name)

	return 0, nil
}

func (h *SysCommonHandler) ReadDirAll(n domain.IOnode, pid uint32) ([]os.FileInfo, error) {

	return nil, nil
}

func (h *SysCommonHandler) GetName() string {
	return h.Name
}

func (h *SysCommonHandler) GetPath() string {
	return h.Path
}

func (h *SysCommonHandler) GetEnabled() bool {
	return h.Enabled
}

func (h *SysCommonHandler) GetType() domain.HandlerType {
	return h.Type
}

func (h *SysCommonHandler) GetService() domain.HandlerService {
	return h.Service
}

func (h *SysCommonHandler) SetEnabled(val bool) {
	h.Enabled = val
}

func (h *SysCommonHandler) SetService(hs domain.HandlerService) {
	h.Service = hs
}