package mine

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ivta/config"
	"ivta/miner"
	"ivta/utils"
)

func Execute() {
	targetURLFlag := flag.String("u", "", "Target URL to mine (required if -tl is not used)")
	targetListFlag := flag.String("tl", "", "Path to a file containing target URLs (required if -u is not used)")
	wordlistFlag := flag.String("w", "wordlist.txt", "Path to the wordlist file (default: wordlist.txt)")
	methodFlag := flag.String("m", "GET", "HTTP method (GET, POST, JSON, XML) (default: GET)")
	headersFlag := flag.String("H", "", "Custom headers (e.g. 'Header1:Value1,Header2:Value2')")
	dataFlag := flag.String("d", "", "Custom data (e.g. 'key1:value1,key2:value2')")
	concurrencyFlag := flag.Int("c", 5, "Number of concurrent requests (default: 5)")
	verboseFlag := flag.Bool("v", false, "Enable verbose mode")
	outputFileFlag := flag.String("o", "miner_results.json", "Output file (default: miner_results.json)")
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

	cfg := config.LoadMinerConfig()

	if *targetURLFlag != "" {
		cfg.TargetURL = *targetURLFlag
	}
	if *targetListFlag != "" {
		cfg.TargetListFile = *targetListFlag
	}
	cfg.WordlistFile = *wordlistFlag
	cfg.Method = *methodFlag
	cfg.Headers = *headersFlag
	cfg.Data = *dataFlag
	cfg.Concurrency = *concurrencyFlag
	cfg.Verbose = *verboseFlag
	cfg.OutputFile = *outputFileFlag

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
					headers[parts[0]] = parts[1]
				}
			}
		}

		data := make(map[string]string)
		if cfg.Data != "" {
			dataPairs := strings.Split(cfg.Data, ",")
			for _, pair := range dataPairs {
				parts := strings.Split(pair, ":")
				if len(parts) == 2 {
					data[parts[0]] = parts[1]
				}
			}
		}

		options := miner.RequestOptions{
			Method:  cfg.Method,
			Headers: headers,
			Data:    data,
		}

		results, err := miner.BruteForce(target, wordlist, options, cfg.Concurrency)
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
	fmt.Println("Discover parameters using brute-force and response analysis.")
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
