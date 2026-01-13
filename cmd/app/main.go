package main

import (
	"flag"

	"github.com/ensoria/projecttemplate/internal/bootstrap"
	"github.com/ensoria/projecttemplate/internal/scheduler"
)

// main entry point

// TODO: メイン関数をcobraに置き換える
// cobraを使うことで、internal/cmd配下に複数のコマンドを配置できるようになる
// cobraをそのままよりも、cmdkitか、cmdgearでラップするか?
func main() {

	// FIXME: configのenvを使って、ここのリストを修正する
	// envList := slices.Join(env.StringList, ", ")
	envVal := flag.String("e", "local", "it must be either [local], [develop], [staging], [production] or [testing].")
	isScheduler := flag.Bool("scheduler", false, "if true, run as scheduler.")
	flag.Parse()
	if envVal == nil {
		panic("Please specify the environment with -e option. It must be either [local], [develop], [staging], [production] or [testing].")
	}

	if isScheduler != nil && *isScheduler {
		scheduler.Start(envVal)
	} else {
		bootstrap.Run(envVal)
	}
}
