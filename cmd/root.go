/*
Copyright © 2020 DrMagPie

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/webhooks.v5/github"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "githook",
		Short: "WebHook app",
		Run: func(cmd *cobra.Command, args []string) {
			if d, e := cmd.Flags().GetBool("debug"); e == nil && d {
				log.SetLevel(log.DebugLevel)
			}
			run()
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/githook/config.yaml)")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	log.Info("Loading Config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		log.Debug("config file ", cfgFile)
	} else {
		home, err := os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		viper.SetConfigFile(fmt.Sprint(home, "/githook/config.yml"))
	}
	viper.SetDefault("token", "94a08da1fecbb6e8b46990538c7b50b2")
	viper.SetDefault("url", "/github")
	viper.SetDefault("command", "echo hello")
	viper.SetDefault("port", "3000")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Error(err)
	}
}

func run() {
	log.Debug("Token: ", viper.GetString("token"))
	hook, err := github.New(github.Options.Secret(viper.GetString("token")))
	if err != nil {
		log.Fatal("Failed to create webhook")
	}

	http.HandleFunc(viper.GetString("url"), func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.PushEvent)
		if err != nil && err == github.ErrEventNotFound {
			log.Error("Event was not present in headdes")
		}
		log.Debug(fmt.Sprintf("%+v", payload))

		switch payload := payload.(type) {
		case github.PushPayload:
			log.Debug(payload.Ref)
			if payload.Ref == "refs/heads/gh-pages" {
				log.Info("Deploing ", payload.HeadCommit.ID)
				out, err := exec.Command("/bin/sh", viper.GetString("command")).Output()
				if err != nil {
					log.Error(err)
					break
				}
				log.Info(string(out))
			}
		default:
			log.Warn("This event is not supported")
		}
	})
	log.Info(fmt.Sprintf("Started gitHook on 0.0.0.0:%s%s", viper.GetString("port"), viper.GetString("url")))
	http.ListenAndServe(fmt.Sprint(":", viper.GetString("port")), nil)
}
