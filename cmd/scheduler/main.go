package main

import (
	"log"

	"github.com/ensoria/projecttemplate/internal/app/bootstrap/scheduler"
	"github.com/spf13/pflag"
)

func main() {
	// FIXME: configのenvを使って、ここのリストを修正する
	// envList := slices.Join(env.StringList, ", ")
	envVal := pflag.StringP("env", "e", "local", "it must be either [local], [develop], [staging], [production] or [testing].")
	pflag.Parse()

	if envVal == nil {
		log.Fatal("Please specify the environment with -e option. It must be either [local], [develop], [staging], [production] or [testing].")
	}

	if err := scheduler.Start(envVal); err != nil {
		log.Fatal(err)
	}
}
