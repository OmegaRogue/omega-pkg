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
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zcalusic/sysinfo"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"omega-pkg/internal/managers"
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
		var si sysinfo.SysInfo
		si.GetSysInfo()
		//data, err := json.MarshalIndent(&si, "", "  ")
		//if err != nil {
		//	log.Fatal().Err(err).Send()
		//}
		//fmt.Println(string(data))
		ctx := log.Logger.WithContext(context.Background())
		ctx = context.WithValue(ctx, lang.DryrunContextKey, viper.GetBool("dryrun"))
		c := viper.Get("config").(lang.Config)
		if err := c.Run(ctx); err != nil {
			log.Fatal().Err(err).Msg("run")
		}
		//data, err = json.MarshalIndent(&c, "", "  ")
		//if err != nil {
		//	log.Fatal().Err(err).Send()
		//}
		//fmt.Println(string(data))

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("run")
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
	rootCmd.Flags().BoolP("dryrun", "d", false, "print commands to run to output")
	err := viper.BindPFlag("dryrun", rootCmd.Flags().Lookup("dryrun"))
	if err != nil {
		log.Fatal().Err(err).Msg("bind flag dryrun")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	////var c lang.Config
	////moreDiags := gohcl.DecodeBody(f.Body, nil, &c)
	parser := hclparse.NewParser()
	base, baseDiags := parser.ParseHCL(managers.Base, "base.hcl")

	f, diags := parser.ParseHCLFile("server.hcl")

	diags = append(diags, baseDiags...)

	body := hcl.MergeBodies([]hcl.Body{base.Body, f.Body})
	var c lang.Config
	ctx, err := lang.BuildGlobalContext()
	if err != nil {
		log.Fatal().Err(err).Msg("Error build global hcl context")
	}

	userfuncs, remain, funcDiags := userfunc.DecodeUserFunctions(body, "func", func() *hcl.EvalContext { return ctx })
	diags = append(diags, funcDiags...)

	ctx.Functions = lo.Assign[string, function.Function](userfuncs, ctx.Functions)

	locals, remain, localDiags := lang.DecodeLocals(remain, ctx)
	diags = append(diags, localDiags...)
	ctx.Variables["local"] = cty.ObjectVal(locals)

	bodyDiags := gohcl.DecodeBody(remain, ctx, &c)
	diags = append(diags, bodyDiags...)

	wr := hcl.NewDiagnosticTextWriter(
		zerolog_extension.LoggerWithLevel(log.Logger, zerolog.ErrorLevel), // writer to send messages to
		parser.Files(), // the parser's file cache, for source snippets
		1000,           // wrapping width
		true,           // generate colored/highlighted output
	)

	validationDiags := c.Validate(ctx)
	diags = append(diags, validationDiags...)

	if err := wr.WriteDiagnostics(diags); err != nil {
		log.Fatal().Err(err).Msg("Error writing diagnostics")
	}
	viper.Set("config", c)
	viper.Set("ctx", ctx)
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
