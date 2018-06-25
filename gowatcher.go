package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"os/exec"

	"github.com/rjeczalik/notify"
)

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
	// never reached
	panic(true)
	return nil, nil
}

func main() {
	fmt.Println("welcome to gowatcher by suyash")
	fmt.Println("started ....")
	var localCmd *exec.Cmd
	var localCmd1 *exec.Cmd
	envArgs := os.Args[1]
	absolutePath := os.Args[2]
	pathtodir := os.Args[3]

	for {
		localCmd = exec.Command("/usr/local/bin/go", "build", absolutePath)
		localCmd.Env = os.Environ()
		parseEnvArgs := strings.Split(envArgs, " ")
		for _, value := range parseEnvArgs {
			localCmd.Env = append(localCmd.Env, value)
		}
		fmt.Println("Started Building")
		if err := localCmd.Start(); err != nil {
			fmt.Println("Command Error:", err)
		}
		procState, procerr := localCmd.Process.Wait()
		if procerr != nil {
			fmt.Println("process state error")
			return
		}
		if procState.Success() == false {
			fmt.Println("Build Failed")
		} else {
			localCmd1 = exec.Command("/usr/local/bin/go", "run", absolutePath)
			//localCmd1.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			localCmd1.Env = os.Environ()
			parseEnvArgs := strings.Split(envArgs, " ")
			for _, value := range parseEnvArgs {
				localCmd1.Env = append(localCmd1.Env, value)
			}

			var stdout, stderr []byte
			var errStdout, errStderr error
			stdoutIn, _ := localCmd1.StdoutPipe()
			stderrIn, _ := localCmd1.StderrPipe()
			fmt.Println("Started Running...")
			if err := localCmd1.Start(); err != nil {
				fmt.Println("Command Error:", err)
			}
			go func() {
				stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
			}()

			go func() {
				stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
			}()

			if errStdout != nil || errStderr != nil {
				log.Fatalf("failed to capture stdout or stderr\n")
			}
		}
		// Make the channel buffered to ensure no event is dropped. Notify will drop
		// an event if the receiver is not able to keep up the sending pace.
		c := make(chan notify.EventInfo, 1)

		// Set up a watchpoint listening on events within current working directory.
		// Dispatch each create and remove events separately to c.
		if err := notify.Watch(pathtodir+"/...", c, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil {
			fmt.Println("Notify Error", err)
		}
		defer notify.Stop(c)

		// Block until an event is received.
		ei := <-c
		if localCmd1 != nil {
			localCmd1.Process.Signal(syscall.SIGQUIT)
			kerr := localCmd1.Process.Kill()
			fmt.Println(kerr)
		}
		log.Println("event triggered :", ei)
	}
}
