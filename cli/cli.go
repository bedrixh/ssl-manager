package cli

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

const ()

type Cli struct {
	Subcommands []Subcommand
	Flags       []Flag
}

type Subcommand struct {
	Name         string
	Aliases      []string
	HelpMessage  string
	CustomValues map[any]any
	Function     func([]Flag)
	Flags        []Flag

	Subcommands []Subcommand
}

type Flag struct {
	Name          string
	Aliases       []string
	Default       any
	HelpMessage   string
	CustomOptions map[any]any

	Value *any
}

func (c *Cli) Parse(Args []string) error {
	if len(Args) < 2 {
		fmt.Errorf("Not enought arguments to parse subcommand and its flags")
	}
	var foundSubcommand *Subcommand = nil

	for _, subcommand := range c.Subcommands {
		if subcommand.Name == Args[1] || slices.Contains(subcommand.Aliases, Args[1]) {
			foundSubcommand = &subcommand
		}
	}

	return nil
}:

func (c *Cli) PrintUsage() error {
	fmt.Printf("Usage: %s [subcommand] [flags]\n\n", os.Args[0])

	fmt.Println("Subcommands:")
	for _, subcommand := range c.Subcommands {
		fmt.Printf("  %s\n\t%s\n", subcommand.Name, subcommand.HelpMessage)
	}
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	return nil
}
