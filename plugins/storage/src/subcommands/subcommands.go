package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/storage"

	flag "github.com/spf13/pflag"
)

func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "default":
		err = storage.CommandHelp()
	case "create":
		args := flag.NewFlagSet("storage:create", flag.ExitOnError)
		scheduler := args.String("scheduler", storage.SchedulerDockerLocal, "--scheduler: target scheduler (docker-local, k3s)")
		size := args.String("size", "", "--size: PVC size (k3s only, e.g. 2Gi)")
		accessMode := args.String("access-mode", "", "--access-mode: PVC access mode (k3s only)")
		storageClass := args.String("storage-class-name", "", "--storage-class-name: PVC storage class (k3s only)")
		namespace := args.String("namespace", "", "--namespace: PVC namespace (k3s only)")
		chown := args.String("chown", "", "--chown: chown option (docker-local only)")
		reclaim := args.String("reclaim-policy", "", "--reclaim-policy: PV reclaim policy (Retain or Delete, k3s only)")
		annotations := args.StringSlice("annotation", nil, "--annotation key=value: PVC annotation (repeatable)")
		labels := args.StringSlice("label", nil, "--label key=value: PVC label (repeatable)")
		args.Parse(os.Args[2:])
		name := args.Arg(0)
		path := args.Arg(1)
		annotMap, parseErr := parseKVPairs(*annotations)
		if parseErr != nil {
			err = parseErr
			break
		}
		labelMap, parseErr := parseKVPairs(*labels)
		if parseErr != nil {
			err = parseErr
			break
		}
		err = storage.CommandCreate(storage.CommandCreateInput{
			Name:          name,
			Path:          path,
			Scheduler:     *scheduler,
			Size:          *size,
			AccessMode:    *accessMode,
			StorageClass:  *storageClass,
			Namespace:     *namespace,
			Chown:         *chown,
			ReclaimPolicy: *reclaim,
			Annotations:   annotMap,
			Labels:        labelMap,
		})
	case "destroy":
		args := flag.NewFlagSet("storage:destroy", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = storage.CommandDestroy(args.Arg(0))
	case "ensure-directory":
		args := flag.NewFlagSet("storage:ensure-directory", flag.ExitOnError)
		chown := args.String("chown", "herokuish", "--chown: chown option (herokuish, heroku, paketo, root, false)")
		args.Parse(os.Args[2:])
		directory := args.Arg(0)
		common.LogWarn("storage:ensure-directory is deprecated; use storage:create instead.")
		err = storage.CommandEnsureDirectory(directory, *chown)
	case "info":
		args := flag.NewFlagSet("storage:info", flag.ExitOnError)
		format := args.String("format", "text", "--format: output format (text, json)")
		args.Parse(os.Args[2:])
		err = storage.CommandInfo(args.Arg(0), *format)
	case "list":
		args := flag.NewFlagSet("storage:list", flag.ExitOnError)
		format := args.String("format", "text", "--format: output format (text, json)")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = storage.CommandList(appName, *format)
	case "list-entries":
		args := flag.NewFlagSet("storage:list-entries", flag.ExitOnError)
		scheduler := args.String("scheduler", "", "--scheduler: filter to a single scheduler")
		format := args.String("format", "text", "--format: output format (text, json)")
		args.Parse(os.Args[2:])
		err = storage.CommandListEntries(*scheduler, *format)
	case "mount":
		args := flag.NewFlagSet("storage:mount", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		mountPath := args.Arg(1)
		err = storage.CommandMount(appName, mountPath)
	case "report":
		args := flag.NewFlagSet("storage:report", flag.ExitOnError)
		format := args.String("format", "stdout", "format: [ stdout | json ]")
		reportArgs, flagErr := common.ParseReportArgs("storage", os.Args[2:])
		if flagErr == nil {
			args.Parse(reportArgs.OSArgs)
			appName := args.Arg(0)
			if reportArgs.IsGlobal {
				appName = "--global"
			}
			err = storage.CommandReport(appName, *format, reportArgs.InfoFlag)
		}
	case "unmount":
		args := flag.NewFlagSet("storage:unmount", flag.ExitOnError)
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		mountPath := args.Arg(1)
		err = storage.CommandUnmount(appName, mountPath)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}

func parseKVPairs(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := map[string]string{}
	for _, pair := range pairs {
		idx := strings.Index(pair, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("expected key=value, got %q", pair)
		}
		out[pair[:idx]] = pair[idx+1:]
	}
	return out, nil
}
