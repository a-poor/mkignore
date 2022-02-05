package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/urfave/cli/v2"
)

const version = "0.1.0"

var compiled = time.Now()

const appDescription = ``

func main() {
	// Define the CLI app
	app := cli.App{
		Name:    "mkignore",
		Version: version,
		Usage:   "",
		Authors: []*cli.Author{{
			Name: "Austin Poor",
		}},
		Compiled:    compiled,
		Description: appDescription,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "include-community",
				Aliases: []string{"c"},
				Usage:   "Include community gitignore templates",
			},
			&cli.BoolFlag{
				Name:    "include-global",
				Aliases: []string{"g"},
				Usage:   "Include global gitignore templates",
			},
			&cli.BoolFlag{
				Name:    "append",
				Aliases: []string{"a"},
				Usage:   "If a .gitignore file already exists, append to it instead of overwriting it.",
			},
			&cli.PathFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Usage:       "Write the .gitignore file to the specified `PATH`",
				DefaultText: ".gitignore",
			},
			&cli.StringSliceFlag{
				Name:    "templates",
				Aliases: []string{"t"},
				Usage: "Specify a list of gitignore templates to use." +
					"If not specified, templates will be selected interactively.",
			},
		},
		Action: func(c *cli.Context) error {
			// Create a struct to hold the options
			// and get CLI flags
			settings := struct {
				IncludeCommunity bool
				IncludeGlobal    bool
				Append           bool
				Path             string
				Templates        []string
			}{
				IncludeCommunity: c.Bool("include-community"),
				IncludeGlobal:    c.Bool("include-global"),
				Append:           c.Bool("append"),
				Path:             c.Path("path"),
				Templates:        c.StringSlice("templates"),
			}

			// Load the templates
			gitignores, err := GetGitignores()
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("Error loading gitignore templates: %s", err),
					1,
				)
			}

			// Get the params set from the CLI.
			// For the rest, add them to the questions.
			if c.IsSet("include-community") {
				settings.IncludeCommunity = c.Bool("include-community")
			} else {
				err = survey.AskOne(
					&survey.Confirm{
						Message: "Include community templates?",
					},
					&settings.IncludeCommunity,
				)
				if err == terminal.InterruptErr {
					return cli.Exit("", 0)
				}
				if err != nil {
					return cli.Exit("Error getting user input.", 1)
				}
			}

			if c.IsSet("include-global") {
				settings.IncludeGlobal = c.Bool("include-global")
			} else {
				err = survey.AskOne(
					&survey.Confirm{
						Message: "Include global templates?",
					},
					&settings.IncludeGlobal,
				)
				if err == terminal.InterruptErr {
					return cli.Exit("", 0)
				}
				if err != nil {
					return cli.Exit("Error getting user input.", 1)
				}
			}

			if c.IsSet("path") {
				settings.Path = c.Path("path")
			} else {
				err = survey.AskOne(
					&survey.Input{
						Message: ".gitignore file path:",
						Default: ".gitignore",
					},
					&settings.Path,
				)
				if err == terminal.InterruptErr {
					return cli.Exit("", 0)
				}
				if err != nil {
					return cli.Exit("Error getting user input.", 1)
				}
			}

			// Is the path a directory? If so, the filename should be .gitingore
			pInfo, err := os.Stat(settings.Path)
			if err != nil && !os.IsNotExist(err) {
				return cli.Exit(
					fmt.Sprintf("Error getting file info: %s", err),
					1,
				)
			}
			if pInfo.IsDir() {
				settings.Path = path.Join(settings.Path, ".gitignore")
			}

			// Does the parent directory exist?
			d, _ := path.Split(settings.Path)
			if _, err := os.Stat(d); d != "" && os.IsNotExist(err) {
				return cli.Exit(
					fmt.Sprintf("Error: directory %s does not exist.", d),
					1,
				)
			}

			// Does the .gitignore file already exist? (if so, ask about append...)
			_, err = os.Stat(settings.Path)
			fileExists := !os.IsNotExist(err)
			if c.IsSet("append") {
				settings.Append = c.Bool("append")
			} else if fileExists {
				err = survey.AskOne(
					&survey.Confirm{
						Message: fmt.Sprintf(
							"Append to existing %q file?",
							settings.Path,
						),
					},
					&settings.Append,
				)
				if err == terminal.InterruptErr {
					return cli.Exit("", 0)
				}
				if err != nil {
					return cli.Exit("Error getting user input.", 1)
				}
			}

			var tmplNames []string
			for _, gi := range gitignores {
				if !settings.IncludeCommunity && gi.IsCommunity() {
					continue
				}
				if !settings.IncludeGlobal && gi.IsGlobal() {
					continue
				}

				tmplNames = append(tmplNames, gi.GetLabel())
			}

			if c.IsSet("templates") {
				settings.Templates = c.StringSlice("templates")
			} else {
				err = survey.AskOne(
					&survey.MultiSelect{
						Message: "Selected gitignore templates:",
						Options: tmplNames,
					},
					&settings.Templates,
				)
				if err == terminal.InterruptErr {
					return cli.Exit("", 0)
				}
				if err != nil {
					return cli.Exit("Error getting user input.", 1)
				}
			}

			var selected []*IgnoreFile
			for _, gi := range gitignores {
				for _, tmpl := range settings.Templates {
					if gi.GetLabel() == tmpl {
						selected = append(selected, &gi)
						break
					}
				}
			}

			b, _ := json.MarshalIndent(selected, "", "  ")
			fmt.Println(string(b))

			// Generate .gitignore template
			fmtTmpl, err := ExecIgnoreTmpl(selected)
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("Error generating .gitignore template: %s", err),
					1,
				)
			}

			f, err := os.OpenFile(settings.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("Error opening file %q: %s", settings.Path, err),
					1,
				)
			}
			defer f.Close()
			_, err = f.WriteString(fmtTmpl)
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("Error writing to file %q: %s", settings.Path, err),
					1,
				)
			}

			return nil
		},

		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
	}

	// Run the app
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
