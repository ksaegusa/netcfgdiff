package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	// 読み込んだパッケージ
	"github.com/ksaegusa/netcfgdiff/pkg/netcfgdiff"
)

func main() {
	var ignoreFlags []string
	var targetFlag string

	var rootCmd = &cobra.Command{
		Use:   "netcfgdiff [running-config] [candidate-config]",
		Short: "A network configuration diff tool",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fileRunning := args[0]
			fileCandidate := args[1]

			// 正規表現のコンパイル
			var ignoreRegexps []*regexp.Regexp
			for _, pattern := range ignoreFlags {
				re, err := regexp.Compile(pattern)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid regex '%s': %v\n", pattern, err)
					os.Exit(1)
				}
				ignoreRegexps = append(ignoreRegexps, re)
			}

			// ▼ 修正: netcfgdiff.ParseFile に変更
			runningNodes, err := netcfgdiff.ParseFile(fileRunning, ignoreRegexps)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing running: %v\n", err)
				os.Exit(1)
			}

			// ▼ 修正: netcfgdiff.ParseFile に変更
			candidateNodes, err := netcfgdiff.ParseFile(fileCandidate, ignoreRegexps)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing candidate: %v\n", err)
				os.Exit(1)
			}

			// ターゲットフィルタリング
			if targetFlag != "" {
				// ▼ 修正: netcfgdiff.FilterNodes に変更
				runningNodes = netcfgdiff.FilterNodes(runningNodes, targetFlag)
				candidateNodes = netcfgdiff.FilterNodes(candidateNodes, targetFlag)

				if len(runningNodes) == 0 && len(candidateNodes) == 0 {
					fmt.Printf("Warning: Target block '%s' not found in either config.\n", targetFlag)
				}
			}

			// 比較実行
			fmt.Println("--- Diff Start ---")
			// ▼ 修正: netcfgdiff.DiffConfig に変更
			netcfgdiff.DiffConfig(os.Stdout, runningNodes, candidateNodes, 0)
			fmt.Println("--- Diff End ---")
		},
	}

	rootCmd.Flags().StringArrayVarP(&ignoreFlags, "ignore", "i", []string{}, "Regex pattern to ignore lines")
	rootCmd.Flags().StringVarP(&targetFlag, "target", "t", "", "Target block prefix to compare")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}