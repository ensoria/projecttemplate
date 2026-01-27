package main

import (
	"github.com/ensoria/projecttemplate/internal/app/bootstrap/scheduler"
	"github.com/ensoria/projecttemplate/internal/app/bootstrap/server"
	"github.com/spf13/pflag"
)

func main() {
	// FIXME: configのenvを使って、ここのリストを修正する
	// envList := slices.Join(env.StringList, ", ")
	envVal := pflag.StringP("env", "e", "local", "it must be either [local], [develop], [staging], [production] or [testing].")
	isScheduler := pflag.BoolP("scheduler", "s", false, "if true, run as scheduler.")
	pflag.Parse()

	if envVal == nil {
		panic("Please specify the environment with -e option. It must be either [local], [develop], [staging], [production] or [testing].")
	}

	if isScheduler != nil && *isScheduler {
		scheduler.Start(envVal)
	} else {
		server.Run(envVal)
	}
}
