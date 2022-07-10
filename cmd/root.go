/*
Copyright © 2022 OmegaRogue <omegarogue@omegavoid.codes>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"omega-pkg/pkg/lang"
	"omega-pkg/pkg/zerolog_extension"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "omega-pkg",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		c := viper.Get("").(lang.Config)
		for _, manager := range c.Managers {
			customManager := c.CustomManagerMap[manager.Name]
			manager.Run(context.TODO(), log.Logger, customManager)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("error on run")
		os.Exit(1)
	}
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.omega-pkg.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	////var c lang.Config
	////moreDiags := gohcl.DecodeBody(f.Body, nil, &c)
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile("server.hcl")
	var c lang.Config
	diags = append(diags, gohcl.DecodeBody(f.Body, nil, &c)...)

	wr := hcl.NewDiagnosticTextWriter(
		zerolog_extension.LoggerWithLevel(log.Logger, zerolog.ErrorLevel), // writer to send messages to
		parser.Files(), // the parser's file cache, for source snippets
		1000,           // wrapping width
		true,           // generate colored/highlighted output
	)

	if err := c.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Error parsing config")
	}

	if err := wr.WriteDiagnostics(diags); err != nil {
		log.Fatal().Err(err).Msg("Error writing diagnostics")
	}
	viper.Set("", c)
	//if cfgFile != "" {
	//	viper.Set(cfgFile)
	//	// Use config file from the flag.
	//	//viper.SetConfigFile(cfgFile)
	//} else {
	//	// Find home directory.
	//	home, err := os.UserHomeDir()
	//	cobra.CheckErr(err)
	//}
}
