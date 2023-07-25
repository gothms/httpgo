package command

import (
	"fmt"
	"github.com/gothms/httpgo/framework/cobra"
	"github.com/gothms/httpgo/framework/contract"
)

// DemoCommand helpCommand show current envionment
var DemoCommand = &cobra.Command{
	Use:   "demo",
	Short: "demo for framework",
	Run: func(cmd *cobra.Command, args []string) {
		container := cmd.GetContainer()
		appService := container.MustMake(contract.AppKey).(contract.App)
		fmt.Println("app base folder:", appService.BaseFolder())
	},
}
