package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "interactive loop for multiple commands",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runShell()
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func runShell() {
	fmt.Println("welcome to the trident cli\n")
	fmt.Println(`For help type "campaigns man" or "tasks man"`)
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionSuggestionBGColor(prompt.Black),
		prompt.OptionSuggestionTextColor(prompt.Red),
		prompt.OptionDescriptionBGColor(prompt.LightGray),
		prompt.OptionDescriptionTextColor(prompt.Black),
		prompt.OptionScrollbarThumbColor(prompt.Black),
		prompt.OptionScrollbarBGColor(prompt.Red))
	p.Run()
}

func completer(in prompt.Document) []prompt.Suggest {
	d := in.GetWordBeforeCursor()
	if d == "" {
		return []prompt.Suggest{}
	}
	s := []prompt.Suggest{
		{Text: "agents man", Description: "Agents man page"},
		{Text: "agents list", Description: "List all agents"},
		{Text: "agents config", Description: "Upload config to the server"},
		{Text: "agents build", Description: "Get agent binaries from server"},
		{Text: "agents modify", Description: "Modify config of existing agent"},
		{Text: "agents tasks", Description: "Get the tasks of a specific agent"},
		{Text: "tasks results", Description: "Get the results from a specific task"},
		{Text: "tasks add", Description: "Add a new task to an agent"},
		{Text: "--agent", Description: "Agent ID"},
		{Text: "--os", Description: "Operating System"},
		{Text: "--path", Description: "Download Path"},
		{Text: "--arch", Description: "Architecture"},
		{Text: "--sleep", Description: "Agent sleep time"},
		{Text: "--jitter", Description: "Agent Jitter percentage"},
		{Text: "--group", Description: "Agent Group"},
		{Text: "--file", Description: "Config file path"},
		{Text: "--arg", Description: "Arguments for a task action"},
		{Text: "--action", Description: "Action to be performed by agent"},
		{Text: "--chunk", Description: "Download chunk size"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func executor(in string) {
	if in == "exit" {
		fmt.Print("trident cli shutting down")
		os.Exit(1)
	}

	if in == "" {
		return
	}

	cmd, flags, err := rootCmd.Find(strings.Fields(in))
	if err != nil {
		fmt.Print(err.Error() + "\n")
		return
	}

	cmd.ParseFlags(flags)
	cmd.Run(cmd, flags)
	fmt.Print("\n")
}
