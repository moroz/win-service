package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type handler struct{}

func (h *handler) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	defer func() {
		status <- svc.Status{State: svc.StopPending}
	}()

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Print("Shutting down...")
				return false, 1
			case svc.Pause:
				status <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				log.Printf("Unexpected service control request #%d", c)
			}
		}
	}
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
	err = os.MkdirAll("./log", 0o755)
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

	r := NewRouter()

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running within Windows service: %s", err)
	}

	if inService {
		log.Printf("Starting service %s", SERVICE_NAME)
		go func() {
			log.Fatal(http.ListenAndServe(":3000", r))
		}()
		runService(SERVICE_NAME, false)
	}

	log.Fatal(http.ListenAndServe(":3000", r))
}
