package command

import (
	"fmt"
	"github.com/gothms/httpgo/framework/cobra"
	"github.com/gothms/httpgo/framework/contract"
)

func initEnvCommand() *cobra.Command {
	//envCommand.AddCommand(envListCommand)
	return envCommand
}

var envCommand = &cobra.Command{
	Use:   "env",
	Short: "获取当前 App 环境",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取 env 环境
		container := cmd.GetContainer()
		//fmt.Printf("env type %T\n", container.MustMake(contract.EnvKey))
		envServic := container.MustMake(contract.EnvKey).(contract.Env)
		// 打印环境
		fmt.Println("environment:", envServic.AppEnv())
		//fmt.Println("environment:", envServic.All())
	},
}
var envListCommand = &cobra.Command{
	//	TODO
}
