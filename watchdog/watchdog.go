// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package watchdog

import (
	"os"
	"os/exec"
	"time"

	"github.com/keys-pub/updater/process"
)

// ExitOn describes when a program should exit (not-restart)
type ExitOn string

const (
	// ExitOnNone means the program should always be restarted
	ExitOnNone ExitOn = ""
	// ExitOnSuccess means the program should only restart if errored
	ExitOnSuccess ExitOn = "success"
	// ExitAllOnSuccess means the program should only restart if errored,
	// otherwise exit this watchdog. Intended for Windows
	ExitAllOnSuccess ExitOn = "all"
)

// Program is a program at path with arguments
type Program struct {
	Path   string
	Args   []string
	ExitOn ExitOn
}

// Watch monitors programs and restarts them if they aren't running
func Watch(programs []Program, restartDelay time.Duration) error {
	// Terminate any existing programs that we are supposed to monitor
	logger.Infof("Terminating any existing programs we will be monitoring")
	terminateExisting(programs)

	// any program can terminate everything if it's ExitAllOnSuccess
	exitAll := func() {
		logger.Infof("Terminating any other programs we are monitoring")
		terminateExisting(programs)
		os.Exit(0)
	}
	// Start monitoring all the programs
	watchPrograms(programs, restartDelay, exitAll)

	return nil
}

func includesARealProcess(pids []int) bool {
	for _, p := range pids {
		if p > 0 {
			return true
		}
	}
	return false
}

// terminateExisting sends a kill signal to every pid that matches the executables of the
// programs. It then loops and tries to send these kill signals again to mitigate a race
// condition where another instance of the watchdog can start another instance of a program
// while this instance of the watchdog is sending its kill signals.
func terminateExisting(programs []Program) {
	logger.Infof("Terminate existing programs")
	var killedPids []int
	for i := 1; i <= 3; i++ {
		killedPids = sendKillToPrograms(programs)
		if !includesARealProcess(killedPids) {
			logger.Infof("none of these programs are running")
			return
		}
		logger.Infof("Terminated pids %v", killedPids)
		time.Sleep(200 * time.Millisecond)
	}
}

func sendKillToPrograms(programs []Program) (killedPids []int) {
	// Terminate any monitored processes
	// this logic also exists in the updater, so if you want to change it, look there too.
	ospid := os.Getpid()
	for _, program := range programs {
		matcher := process.NewMatcher(program.Path, process.PathEqual)
		matcher.ExceptPID(ospid)
		logger.Infof("Terminating %s", program.Path)
		pids := process.TerminateAll(matcher, time.Second)
		killedPids = append(killedPids, pids...)
	}
	return killedPids
}

func watchPrograms(programs []Program, delay time.Duration, exitAll func()) {
	for _, program := range programs {
		go watchProgram(program, delay, exitAll)
	}
}

// watchProgram will monitor a program and restart it if it exits.
// This method will run forever.
func watchProgram(program Program, restartDelay time.Duration, exitAll func()) {
	for {
		start := time.Now()
		logger.Infof("Starting %#v", program)
		cmd := exec.Command(program.Path, program.Args...)
		err := cmd.Run()
		if err != nil {
			logger.Errorf("Error running program: %q; %s", program, err)
		} else {
			logger.Infof("Program finished: %q", program)
			if program.ExitOn == ExitOnSuccess {
				logger.Infof("Program configured to exit on success, not restarting")
				break
			} else if program.ExitOn == ExitAllOnSuccess {
				logger.Infof("Program configured to exit on success, exiting")
				exitAll()
			}
		}
		logger.Infof("Program ran for %s", time.Since(start))
		if time.Since(start) < restartDelay {
			logger.Infof("Waiting %s before trying to start command again", restartDelay)
			time.Sleep(restartDelay)
		}
	}
}
