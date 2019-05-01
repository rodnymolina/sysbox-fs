package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nestybox/sysvisor/sysvisor-fs/domain"
	"github.com/nestybox/sysvisor/sysvisor-fs/fuse"
	"github.com/nestybox/sysvisor/sysvisor-fs/handler"
	"github.com/nestybox/sysvisor/sysvisor-fs/handler/implementations"
	"github.com/nestybox/sysvisor/sysvisor-fs/ipc"
	"github.com/nestybox/sysvisor/sysvisor-fs/state"
	"github.com/nestybox/sysvisor/sysvisor-fs/sysio"
)

// TODO: Beautify me please.
func usage() {
	fmt.Fprintf(os.Stderr, "Usafe of %s\n", os.Args[0])
	fmt.Fprintf(os.Stderr, " %s <file-system mount-point>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n Example: sysvisorfs /var/lib/sysvisorfs\n\n")
	fmt.Fprintf(os.Stderr, "OtherFUSE options:\n")
	flag.PrintDefaults()
}

//
// Sysvisorfs signal handler goroutine.
//
func signalHandler(signalChan chan os.Signal, fs domain.FuseService) {

	s := <-signalChan

	switch s {

	// TODO: Handle SIGHUP differently -- e.g. re-read sysvisorfs conf file
	case syscall.SIGHUP:
		log.Println("Sysvisorfs caught signal: SIGHUP")

	case syscall.SIGSEGV:
		log.Println("Sysvisorfs caught signal: SIGSEGV")

	case syscall.SIGINT:
		log.Println("Sysvisorfs caught signal: SIGTINT")

	case syscall.SIGTERM:
		log.Println("Sysvisorfs caught signal: SIGTERM")

	case syscall.SIGQUIT:
		log.Println("Sysvisorfs caught signal: SIGQUIT")

	default:
		log.Println("Sysvisorfs caught unknown signal")
	}

	log.Println("Unmounting sysvisorfs from mountpoint", fs.MountPoint(), "Exitting...")
	fs.Unmount()

	// Deferring exit() to allow FUSE to dump unnmount() logs
	time.Sleep(2)

	os.Exit(0)
}

//
// Sysvisor-fs main function
//
func main() {

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(-1)
	}

	// TODO: Enhance cli/parsing logic.
	if flag.Arg(0) == "nsenter" {
		implementations.Nsenter()
		return
	}

	// Sysvisor-fs mountpoint.
	mountPoint := flag.Arg(0)

	//
	// Initiate sysvisor-fs' services.
	//

	var containerStateService = state.NewContainerStateService()

	var handlerService = handler.NewHandlerService(
		handler.DefaultHandlers,
		containerStateService)

	var ioService = sysio.NewIOService(sysio.IOFileService)

	var ipcService = ipc.NewIpcService(containerStateService, ioService)
	ipcService.Init()

	var fuseService = fuse.NewFuseService(
		"/",
		mountPoint,
		ioService,
		handlerService)

	// Launch signal-handler to ensure mountpoint is properly unmounted
	// during shutdown.
	var signalChan = make(chan os.Signal)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGSEGV,
		syscall.SIGQUIT)
	go signalHandler(signalChan, fuseService)

	// Initiate sysvisor-fs' FUSE service.
	if err := fuseService.Run(); err != nil {
		log.Fatal(err)
	}
}