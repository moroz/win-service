package main

import (
	"log"
	"os"
	"path"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type handler struct{}

func (h *handler) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	tick := time.NewTicker(10 * time.Second)
	// tick := time.Tick(10 * time.Second)
	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case <-tick.C:
			log.Print("Tick handled!")
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Print("Shutting down...")
				break loop
			case svc.Pause:
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick.Stop()
			case svc.Continue:
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = time.NewTicker(10 * time.Second)
			default:
				log.Printf("Unexpected service control request #%d", c)
			}
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 1
}

func runService(name string, isDebug bool) {
	if isDebug {
		err := debug.Run(name, &handler{})
		if err != nil {
			log.Fatalf("Error running service %s in debug mode: %s", name, err)
		}
	} else {
		err := svc.Run(name, &handler{})
		if err != nil {
			log.Fatalf("Error running service %s in Service Control mode: %s", name, err)
		}
	}
}

var DEBUG = os.Getenv("DEBUG") != ""

const SERVICE_NAME = "moroz-winservice"

func setupLog() *os.File {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	logFile := path.Join(pwd, `log/debug.log`)

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("error opening file: %s", err)
	}
	log.SetOutput(f)
	return f
}

func main() {
	f := setupLog()
	defer f.Close()

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running within Windows service: %s", err)
	}
	if inService {
		runService(SERVICE_NAME, false)
		return
	}
}
