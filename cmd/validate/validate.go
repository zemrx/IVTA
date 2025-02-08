package cmd

import (
	"flag"
	"fmt"
	"os"

	"ivta/config"
	"ivta/utils"
	"ivta/validator"
)

func Execute() {
	targetURLFlag := flag.String("u", "", "Target URL to validate (required if -tl is not used)")
	targetListFlag := flag.String("tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	outputFileFlag := flag.String("o", "validator_results.json", "Path to the output file (default: validator_results.json)")
	verboseFlag := flag.Bool("v", false, "Enable verbose mode")
	helpFlag := flag.Bool("h", false, "Display this help message")

	flag.Parse()

	if *helpFlag {
		Help()
		os.Exit(0)
	}

	if *targetURLFlag == "" && *targetListFlag == "" {
		fmt.Println("Error: You must provide either a target URL (-u) or a target list file (-tl).")
		Help()
		os.Exit(1)
	}

	cfg := config.LoadValidatorConfig()
	cfg.OutputFile = *outputFileFlag
	cfg.Verbose = *verboseFlag

	var targets []string
	if *targetListFlag != "" {
		var err error
		targets, err = utils.ReadTargetList(*targetListFlag)
		if err != nil {
			fmt.Println("Error reading target list:", err)
			os.Exit(1)
		}
	} else {
		targets = []string{*targetURLFlag}
	}

	var resultParams []string
	for _, target := range targets {
		fmt.Println("Processing target:", target)
		reflectedParams, err := validator.IdentifyReflectedParams(target)
		if err != nil {
			fmt.Println("Error processing target:", err)
			continue
		}

		if len(reflectedParams) == 0 {
			fmt.Println("No reflected parameters found.")
			continue
		}

		fmt.Printf("Found %d reflected parameters.\n", len(reflectedParams))
		resultParams = append(resultParams, reflectedParams...)
	}

	config.SaveResults(cfg.OutputFile, nil, nil, nil, nil, resultParams)
	fmt.Println("Results saved to", cfg.OutputFile)
}

func Help() {
	fmt.Println("Usage: ivta validator -u <target_url> [options]")
	fmt.Println("       ivta validator -tl <target_list_file> [options]")
	fmt.Println()
	fmt.Println("Validate URL parameters for reflection vulnerabilities.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required if -tl is not used)")
	fmt.Println("  -tl      Path to a file containing target URLs (required if -u is not used)")
	fmt.Println("  -o       Output file (default: validator_results.json)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -h       Display this help message")
}
