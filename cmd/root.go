/*
Copyright © 2021 myback.space <git@myback.space>

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
	dockerPull "github.com/myback/go-docker-pull"
	"github.com/myback/go-docker-pull/archive"
	"github.com/spf13/cobra"
	"os"
)

var (
	//verbose                      int
	saveCache, onlyDownload      bool
	arch, osType, user, password string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "docker-pull image [image ...]",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Usage()
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		rClient := dockerPull.RegistryClient{
			Arch:     arch,
			OS:       osType,
			Login:    user,
			Password: password,
		}

		for _, img := range args {
			req, err := dockerPull.ParseRequestedImage(img)
			if err != nil {
				fmt.Printf("%s: %s\n", img, err)
				os.Exit(2)
			}

			if err := rClient.Pull(req); err != nil {
				fmt.Printf("%s: %s\n", img, err)
				os.Exit(2)
			}

			if onlyDownload {
				os.Exit(0)
			}

			if err := archive.Tar(req.TempDir(), req.OutputImageName()); err != nil {
				fmt.Println(err)
				os.Exit(2)
			}

			if !saveCache {
				if err := os.RemoveAll(req.TempDir()); err != nil {
					fmt.Println(err)
					os.Exit(2)
				}
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&saveCache, "save-cache", "s", false, "Do not delete the temp folder")
	rootCmd.PersistentFlags().BoolVarP(&onlyDownload, "only-download", "d", false, "Only download layers")
	//rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "")
	rootCmd.PersistentFlags().StringVarP(&arch, "arch", "a", "amd64", "CPU architecture platform image")
	rootCmd.PersistentFlags().StringVarP(&osType, "os", "o", "linux", "OS platform image")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Registry user")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Registry password")
}
