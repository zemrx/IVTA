package mine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"ivta/config"
	"ivta/miner"
	"ivta/utils"
)

func Execute() {
	cfg := config.LoadMinerConfig()

	var targets []string
	if cfg.TargetListFile != "" {
		var err error
		targets, err = utils.ReadTargetList(cfg.TargetListFile)
		if err != nil {
			fmt.Println("Error reading target list:", err)
			os.Exit(1)
		}
	} else {
		targets = []string{cfg.TargetURL}
	}

	wordlist, err := utils.ReadWordlist(cfg.WordlistFile)
	if err != nil {
		fmt.Printf("Error reading wordlist: %v\n", err)
		os.Exit(1)
	}

	for _, target := range targets {
		fmt.Println("Processing target:", target)

		headers := make(map[string]string)
		if cfg.Headers != "" {
			headerPairs := strings.Split(cfg.Headers, ",")
			for _, pair := range headerPairs {
				parts := strings.Split(pair, ":")
				if len(parts) == 2 {
					headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		data := make(map[string]string)
		if cfg.Data != "" {
			dataPairs := strings.Split(cfg.Data, ",")
			for _, pair := range dataPairs {
				parts := strings.Split(pair, ":")
				if len(parts) == 2 {
					data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		options := miner.RequestOptions{
			Method:  cfg.Method,
			Headers: headers,
			Data:    data,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		baselineResp, err := miner.DoRequest(ctx, target, options)
		if err != nil {
			fmt.Printf("Error making baseline request: %v\n", err)
			continue
		}

		extractedParams, wordsExist, err := miner.ExtractParamsFromBaseline(baselineResp, wordlist)
		if err != nil {
			fmt.Printf("Error extracting parameters: %v\n", err)
		} else if wordsExist {
			fmt.Println("Discovered parameters from baseline response:")
			for _, param := range extractedParams {
				fmt.Println(" -", param)
			}
			wordlist = utils.MergeUnique(extractedParams, wordlist)
		}

		results, err := miner.BruteForce(ctx, target, wordlist, options, cfg.Concurrency)
		if err != nil {
			fmt.Printf("Error during brute-force: %v\n", err)
			continue
		}

		fmt.Printf("Discovered %d parameters:\n", len(results))
		for _, param := range results {
			fmt.Println(param)
		}

		config.SaveResults(cfg.OutputFile, nil, nil, nil, nil, results)
		fmt.Println("Results saved to", cfg.OutputFile)
	}
}

func Help() {
	fmt.Println("Usage: ivta.exe miner -u <target_url> [options]")
	fmt.Println("       ivta.exe miner -tl <target_list_file> [options]")
	fmt.Println()
	fmt.Println("Discover parameters using brute-force, response analysis, and baseline extraction.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required if -tl is not used)")
	fmt.Println("  -tl      Path to a file containing target URLs (required if -u is not used)")
	fmt.Println("  -w       Path to the wordlist file (default: wordlist.txt)")
	fmt.Println("  -m       HTTP method (GET, POST, JSON, XML) (default: GET)")
	fmt.Println("  -H       Custom headers (e.g., 'Header1:Value1,Header2:Value2')")
	fmt.Println("  -d       Custom data (e.g., 'key1:value1,key2:value2')")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Output file (default: miner_results.json)")
	fmt.Println("  -h       Display this help message")
}
