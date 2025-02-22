package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc/mgr"
)

func executableExists(p string) bool {
	stat, err := os.Stat(p)
	return err == nil && !stat.Mode().IsDir()
}

func executablePath() (string, error) {
	p, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	if executableExists(p) {
		return p, nil
	}
	if filepath.Ext(p) == "" && executableExists(p+".exe") {
		return p + ".exe", nil
	}
	return "", fmt.Errorf("Failed to determine executable path")
}

func installService(name, desc string) error {
	exec, err := executablePath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	s, err = m.CreateService(name, exec, mgr.Config{DisplayName: desc})
	if err != nil {
		return err
	}
	s.Close()
	return nil
}
