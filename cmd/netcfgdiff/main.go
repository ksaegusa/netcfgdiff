package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ksaegusa/netcfgdiff/pkg/netcfgdiff"
	"github.com/spf13/cobra"
)

func main() {
	var ignoreFlags []string
	var replaceFlags []string
	var profileFlag string
	var targetFlag string

	var rootCmd = &cobra.Command{
		Use:   "netcfgdiff [running-config] [candidate-config]",
		Short: "A network configuration diff tool",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fileRunning := args[0]
			fileCandidate := args[1]

			var ignoreSources []string
			var replaceSources []netcfgdiff.ReplaceRule

			if profileFlag != "" {
				profile, err := netcfgdiff.LoadProfile(profileFlag)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
					os.Exit(1)
				}
				ignoreSources = append(ignoreSources, profile.Ignore...)
				replaceSources = append(replaceSources, profile.Replace...)
			}

			ignoreSources = append(ignoreSources, ignoreFlags...)

			for _, raw := range replaceFlags {
				rule, err := parseReplaceFlag(raw)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid replace rule '%s': %v\n", raw, err)
					os.Exit(1)
				}
				replaceSources = append(replaceSources, rule)
			}

			var ignoreRegexps []*regexp.Regexp
			for _, pattern := range ignoreSources {
				re, err := regexp.Compile(pattern)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid regex '%s': %v\n", pattern, err)
					os.Exit(1)
				}
				ignoreRegexps = append(ignoreRegexps, re)
			}

			replaceRules, err := netcfgdiff.CompileReplaceRules(replaceSources)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid replace rules: %v\n", err)
				os.Exit(1)
			}

			options := netcfgdiff.ParseOptions{
				IgnorePatterns: ignoreRegexps,
				ReplaceRules:   replaceRules,
			}

			runningNodes, err := netcfgdiff.ParseFile(fileRunning, options)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing running: %v\n", err)
				os.Exit(1)
			}

			candidateNodes, err := netcfgdiff.ParseFile(fileCandidate, options)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing candidate: %v\n", err)
				os.Exit(1)
			}

			// ターゲットフィルタリング
			if targetFlag != "" {
				runningNodes = netcfgdiff.FilterNodes(runningNodes, targetFlag)
				candidateNodes = netcfgdiff.FilterNodes(candidateNodes, targetFlag)

				if len(runningNodes) == 0 && len(candidateNodes) == 0 {
					fmt.Printf("Warning: Target block '%s' not found in either config.\n", targetFlag)
				}
			}

			// 比較実行
			fmt.Println("--- Diff Start ---")
			netcfgdiff.DiffConfig(os.Stdout, runningNodes, candidateNodes, 0)
			fmt.Println("--- Diff End ---")
		},
	}

	rootCmd.Flags().StringArrayVarP(&ignoreFlags, "ignore", "i", []string{}, "Regex pattern to ignore lines")
	rootCmd.Flags().StringArrayVarP(&replaceFlags, "replace", "r", []string{}, "Regex replacement rule in pattern=replacement form")
	rootCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "YAML profile with ignore and replace rules")
	rootCmd.Flags().StringVarP(&targetFlag, "target", "t", "", "Target block prefix to compare")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseReplaceFlag(raw string) (netcfgdiff.ReplaceRule, error) {
	pattern, replacement, found := strings.Cut(raw, "=")
	if !found || pattern == "" {
		return netcfgdiff.ReplaceRule{}, fmt.Errorf("expected pattern=replacement")
	}
	return netcfgdiff.ReplaceRule{
		Pattern:     pattern,
		Replacement: replacement,
	}, nil
}
