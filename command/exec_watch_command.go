package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"github.com/coreos/go-etcd/etcd"
	"github.com/joshi4/cobra"
)

var execWatchCmd *cobra.Command

//flags

var execRecursiveFlag bool
var execAfterIndexFlag int

func init() {
	execWatchCmd = &cobra.Command{
		Use:   "exec-watch",
		Short: "watch a key for changes and run an executable",
		Run: func(cmd *cobra.Command, args []string) {
			handleKey(cmd, args, execWatchCommandFunc)
		},
	}

	execWatchCmd.Flags().BoolVar(&execRecursiveFlag, "recursive", false, "watch all values for key and child keys")
	execWatchCmd.Flags().IntVar(&execAfterIndexFlag, "after-index", 0, "watch after the given index")
}

func ExecWatchCommand() *cobra.Command {
	return execWatchCmd
}

// execWatchCommandFunc executes the "exec-watch" command.
func execWatchCommandFunc(cmd *cobra.Command, args []string, client *etcd.Client) (*etcd.Response, error) {
	// _ = io.Copy
	// _ = exec.Command
	// args := c.Args()
	argsLen := len(args)

	if argsLen < 2 {
		return nil, errors.New("Key and command to exec required")
	}

	// key := args[argsLen-1]
	key := args[0]
	cmdArgs := args[1:]

	index := 0
	if execAfterIndexFlag != 0 {
		index = execAfterIndexFlag + 1

	}

	recursive := execRecursiveFlag

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	stop := make(chan bool)

	go func() {
		<-sigch
		stop <- true
		os.Exit(0)
	}()

	receiver := make(chan *etcd.Response)
	client.SetConsistency(etcd.WEAK_CONSISTENCY)
	go client.Watch(key, uint64(index), recursive, receiver, stop)

	for {
		resp := <-receiver
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Env = environResponse(resp, os.Environ())

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
		cmd.Wait()
	}

	return nil, nil
}

func environResponse(resp *etcd.Response, env []string) []string {
	env = append(env, "ETCD_WATCH_ACTION="+resp.Action)
	env = append(env, "ETCD_WATCH_MODIFIED_INDEX="+fmt.Sprintf("%d", resp.Node.ModifiedIndex))
	env = append(env, "ETCD_WATCH_KEY="+resp.Node.Key)
	env = append(env, "ETCD_WATCH_VALUE="+resp.Node.Value)
	return env
}
