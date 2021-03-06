package command

import (
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"github.com/joshi4/cobra"
)

var rmCmd *cobra.Command

// flag variables
var rmDirFlag bool
var rmRecursiveFlag bool
var rmWithValueFlag string
var rmWithIndexFlag int

func init() {

	rmCmd = &cobra.Command{
		Use:   "rm",
		Short: "remove a key",

		Run: func(cmd *cobra.Command, args []string) {
			handleAll(cmd, args, removeCommandFunc)
		},
	}

	rmCmd.Flags().BoolVar(&rmDirFlag, "dir", false, "removes the key if it is an empty directory or a key-value pair")
	rmCmd.Flags().BoolVar(&rmRecursiveFlag, "recursive", false, "removes the key and all child keys(if it is a directory)")
	rmCmd.Flags().StringVar(&rmWithValueFlag, "with-value", "", "previous value")
	rmCmd.Flags().IntVar(&rmWithIndexFlag, "with-index", 0, "previous index")

}

// RemoveCommand returns the *cobra.Command for "rm".
func RemoveCommand() *cobra.Command {
	return rmCmd
}

// removeCommandFunc executes the "rm" command.
func removeCommandFunc(cmd *cobra.Command, args []string, client *etcd.Client) (*etcd.Response, error) {
	if len(args) == 0 {
		return nil, errors.New("Key required")
	}
	key := args[0]
	recursive := rmRecursiveFlag
	dir := rmDirFlag

	// TODO: distinguish with flag is not set and empty flag
	// the cli pkg need to provide this feature
	prevValue := rmWithValueFlag
	prevIndex := uint64(rmWithIndexFlag)

	if prevValue != "" || prevIndex != 0 {
		return client.CompareAndDelete(key, prevValue, prevIndex)
	}

	if recursive || !dir {
		return client.Delete(key, recursive)
	}

	return client.DeleteDir(key)
}
