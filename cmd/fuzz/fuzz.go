package fuzz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"ivta/config"
	"ivta/fuzzer"
	"ivta/utils"
)

func Execute() {
	cfg := config.LoadFuzzConfig()

	if cfg.TargetURL == "" && cfg.TargetListFile == "" {
		fmt.Println("Error: You must provide either a target URL (-u) or a target list file (-tl).")
		Help()
		os.Exit(1)
	}

	if cfg.UseRaft {
		raftFile := "raft.txt"
		if _, err := os.Stat(raftFile); os.IsNotExist(err) {
			fmt.Println("Raft wordlist not found. Downloading...")
			url := "https://gitlab.com/Md_Shaman/SecLists/-/raw/eee1651de7906112719066540ca2c5bf688cf9f2/Discovery/Web-Content/raft-small-directories-lowercase.txt"
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error downloading Raft wordlist: %v\n", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			out, err := os.Create(raftFile)
			if err != nil {
				fmt.Printf("Error creating Raft wordlist file: %v\n", err)
				os.Exit(1)
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				fmt.Printf("Error saving Raft wordlist: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Raft wordlist downloaded successfully.")
		}
		cfg.DirWordlistFile = raftFile
	}

	var bs []int
	if cfg.BlacklistStatusCodes != "" {
		bs = utils.ParseIntSlice(cfg.BlacklistStatusCodes)
	}
	var bl []int
	if cfg.BlacklistLengths != "" {
		bl = utils.ParseIntSlice(cfg.BlacklistLengths)
	}
	var bw []int
	if cfg.BlacklistWordCounts != "" {
		bw = utils.ParseIntSlice(cfg.BlacklistWordCounts)
	}
	var blc []int
	if cfg.BlacklistLineCounts != "" {
		blc = utils.ParseIntSlice(cfg.BlacklistLineCounts)
	}
	var bsw []string
	if cfg.BlacklistSearchWords != "" {
		bsw = strings.Split(cfg.BlacklistSearchWords, ",")
	}
	var br []string
	if cfg.BlacklistRegex != "" {
		br = strings.Split(cfg.BlacklistRegex, ",")
	}
	var extensions []string
	if cfg.Extensions != "" {
		extensions = strings.Split(cfg.Extensions, ",")
	}

	options := fuzzer.FuzzOptions{
		Depth:                cfg.MaxDepth,
		BlacklistStatusCodes: bs,
		BlacklistLengths:     bl,
		BlacklistWordCounts:  bw,
		BlacklistLineCounts:  blc,
		BlacklistSearchWords: bsw,
		BlacklistRegex:       br,
		Extensions:           extensions,
		UserAgent:            cfg.UserAgent,
		Verbose:              cfg.Verbose,
	}

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

	for _, target := range targets {
		fmt.Println("Processing target:", target)
		results := fuzzer.FuzzDirectories(context.Background(), target, cfg.DirWordlistFile, cfg.Concurrency, options)
		fmt.Printf("Found %d valid paths.\n", len(results))

		var resultURLs []string
		for _, res := range results {
			resultURLs = append(resultURLs, res.URL)
		}

		config.SaveResults(cfg.OutputFile, nil, nil, nil, resultURLs, nil)
		fmt.Println("Results saved to", cfg.OutputFile)
	}
}

func Help() {
	fmt.Println("Usage: ivta fuzz -u <target_url> [options]")
	fmt.Println("       ivta fuzz -tl <target_list_file> [options]")
	fmt.Println()
	fmt.Println("Fuzz a website for directories using blacklist filtering.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required if -tl is not used)")
	fmt.Println("  -tl      Path to a file containing target URLs (required if -u is not used)")
	fmt.Println("  -w       Path to the directory wordlist file (default: wordlist.txt)")
	fmt.Println("  -d       Maximum recursion depth for fuzzing (default: 2)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Output file (default: fuzz_results.json)")
	fmt.Println("  -bs      Comma-separated list of blacklisted status codes (e.g. 404,500)")
	fmt.Println("  -bl      Comma-separated list of blacklisted body lengths (in bytes)")
	fmt.Println("  -bw      Comma-separated list of blacklisted word counts")
	fmt.Println("  -blc     Comma-separated list of blacklisted non-empty line counts")
	fmt.Println("  -bsw     Comma-separated list of blacklisted search words")
	fmt.Println("  -br      Comma-separated list of blacklisted regex patterns")
	fmt.Println("  -e       Comma-separated list of extensions to append to words (e.g. php,html)")
	fmt.Println("  -ua      User-Agent string to use")
	fmt.Println("  -raft    Use Raft wordlist if no wordlist is provided")
	fmt.Println("  -h       Display this help message")
}
