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
	case "set":
		args := flag.NewFlagSet("storage:set", flag.ExitOnError)
		size := args.String("size", "", "--size: new PVC size (k3s)")
		accessMode := args.String("access-mode", "", "--access-mode: existing access mode (must match)")
		storageClass := args.String("storage-class-name", "", "--storage-class-name: existing storage class (must match)")
		namespace := args.String("namespace", "", "--namespace: new namespace")
		chown := args.String("chown", "", "--chown: chown option")
		reclaim := args.String("reclaim-policy", "", "--reclaim-policy: PV reclaim policy")
		annotations := args.StringSlice("annotation", nil, "--annotation key=value: PVC annotation (repeatable, replaces all)")
		labels := args.StringSlice("label", nil, "--label key=value: PVC label (repeatable, replaces all)")
		args.Parse(os.Args[2:])
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
		err = storage.CommandSet(storage.CommandSetInput{
			Name:          args.Arg(0),
			Size:          *size,
			AccessMode:    *accessMode,
			StorageClass:  *storageClass,
			Namespace:     *namespace,
			Chown:         *chown,
			ReclaimPolicy: *reclaim,
			Annotations:   annotMap,
			Labels:        labelMap,
		})
	case "exec":
		args := flag.NewFlagSet("storage:exec", flag.ExitOnError)
		image := args.String("image", "", "--image: container image to use (default alpine:3)")
		asUser := args.String("as-user", "", "--as-user: numeric uid to run the exec container as (overrides the entry's chown)")
		args.Parse(os.Args[2:])
		positional := args.Args()
		if len(positional) == 0 {
			err = fmt.Errorf("storage:exec requires a storage entry name")
			break
		}
		name := positional[0]
		var cmd []string
		if len(positional) > 1 {
			cmd = positional[1:]
		}
		err = storage.CommandExec(storage.CommandExecInput{
			Name:   name,
			Image:  *image,
			AsUser: *asUser,
			Args:   cmd,
		})
	case "migrate":
		args := flag.NewFlagSet("storage:migrate", flag.ExitOnError)
		all := args.Bool("all", false, "--all: re-migrate every app on this install")
		args.Parse(os.Args[2:])
		err = storage.CommandMigrate(args.Arg(0), *all)
	case "wait":
		args := flag.NewFlagSet("storage:wait", flag.ExitOnError)
		args.Parse(os.Args[2:])
		err = storage.CommandWait(args.Arg(0))
	case "list-entries":
		args := flag.NewFlagSet("storage:list-entries", flag.ExitOnError)
		scheduler := args.String("scheduler", "", "--scheduler: filter to a single scheduler")
		format := args.String("format", "text", "--format: output format (text, json)")
		args.Parse(os.Args[2:])
		err = storage.CommandListEntries(*scheduler, *format)
	case "mount":
		args := flag.NewFlagSet("storage:mount", flag.ExitOnError)
		containerDir := args.String("container-dir", "", "--container-dir: container path (named-entry form)")
		phases := args.StringSlice("phase", nil, "--phase: phase to mount in (deploy, run; default both)")
		processType := args.String("process-type", "", "--process-type: process type to mount for")
		subpath := args.String("volume-subpath", "", "--volume-subpath: subpath within the entry")
		readonly := args.Bool("volume-readonly", false, "--volume-readonly: mount the volume read-only")
		volumeChown := args.String("volume-chown", "", "--volume-chown: chown option applied at mount time")
		args.Parse(os.Args[2:])
		err = storage.CommandMount(storage.CommandMountInput{
			AppName:      args.Arg(0),
			NameOrPath:   args.Arg(1),
			ContainerDir: *containerDir,
			Phases:       *phases,
			ProcessType:  *processType,
			Subpath:      *subpath,
			Readonly:     *readonly,
			VolumeChown:  *volumeChown,
		})
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
		containerDir := args.String("container-dir", "", "--container-dir: container path (named-entry form, disambiguates duplicates)")
		args.Parse(os.Args[2:])
		err = storage.CommandUnmount(storage.CommandUnmountInput{
			AppName:      args.Arg(0),
			NameOrPath:   args.Arg(1),
			ContainerDir: *containerDir,
		})
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
