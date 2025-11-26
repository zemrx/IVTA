# IVTA - Intelligent Vulnerability Testing & Analysis

A web application security testing tool for discovering hidden paths, parameters, and vulnerabilities.

## What It Does

IVTA combines multiple discovery techniques to find:
- Hidden directories and files
- URL parameters
- Reflected input and unfiltered symbols
- API endpoints

## Modules

### 1. Fuzz - Directory/File Fuzzing
Brute-forces directories and files using wordlists.

**Usage:**
```bash
ivta fuzz -u <url> [options]
```

**Options:**
```
-u <url>           Target URL
-tl <file>         Target list file
-w <file>          Wordlist file
-raft              Use Raft wordlist (auto-downloads)
-e <ext>           Extensions (e.g., php,html,asp)
-ua <string>       User-Agent
-d <depth>         Recursion depth (default: 2)
-c <num>           Concurrency (default: 5)
-v                 Verbose mode
-o <file>          Output file
-bs <codes>        Blacklist status codes
-bl <lengths>      Blacklist content lengths
-bw <counts>       Blacklist word counts
-blc <counts>      Blacklist line counts
-bsw <words>       Blacklist search words
-br <patterns>     Blacklist regex patterns
```

**Example:**
```bash
ivta fuzz -u http://target.com -raft -e php,html -d 2
```

### 2. Hybrid - Combined Discovery
Combines crawling, fuzzing, and parameter mining.

**What it does:**
1. Parses sitemap.xml
2. Crawls HTML links
3. Renders JavaScript to find links
4. Fuzzes discovered paths
5. Mines parameters

**Usage:**
```bash
ivta hybrid -u <url> [options]
```

**Options:**
```
-u <url>           Target URL
-tl <file>         Target list file
-w <file>          Directory wordlist
-p <file>          Parameter wordlist
-raft              Use Raft wordlist
-e <ext>           Extensions
-ua <string>       User-Agent
-d <depth>         Recursion depth
-c <num>           Concurrency
-H <headers>       Custom headers (e.g., "Auth:token")
-ddata <data>      Custom data (e.g., "key:value")
-m <method>        HTTP method (GET, POST, JSON, XML)
-v                 Verbose mode
-o <file>          Output file
```

**Example:**
```bash
ivta hybrid -u http://target.com -raft -e php -H "Authorization:Bearer token"
```

### 3. Miner - Parameter Discovery
Discovers hidden GET/POST parameters.

**Usage:**
```bash
ivta miner -u <url> -w <wordlist> [options]
```

**Options:**
```
-u <url>           Target URL
-tl <file>         Target list file
-w <file>          Parameter wordlist
-m <method>        HTTP method (GET, POST, JSON, XML)
-H <headers>       Custom headers
-d <data>          Custom data
-c <num>           Concurrency
-v                 Verbose mode
-o <file>          Output file
-i <value>         Injection value (default: test-value)
```

**Example:**
```bash
ivta miner -u http://target.com/api -w params.txt -m POST
```

### 4. Validator - Reflection Testing
Tests for reflected parameters and unfiltered symbols.

**Usage:**
```bash
ivta validator -u <url> [options]
```

**Options:**
```
-u <url>           Target URL with parameter
-tl <file>         Target list file
-c <num>           Concurrency (default: 40)
-v                 Verbose mode
-o <file>          Output file
```

**Example:**
```bash
ivta validator -u "http://target.com/search?q=test"
```

### 5. Crawler - Link Discovery
Crawls website to discover links.

**Usage:**
```bash
ivta crawl -u <url> [options]
```

**Options:**
```
-u <url>           Target URL
-tl <file>         Target list file
-d <depth>         Crawl depth (default: 2)
-c <num>           Concurrency (default: 5)
-v                 Verbose mode
-o <file>          Output file
```

**Example:**
```bash
ivta crawl -u http://target.com -d 3
```

## Quick Examples

```bash
# Basic directory fuzzing
ivta fuzz -u http://target.com -raft

# Fuzz with extensions
ivta fuzz -u http://target.com -raft -e php,html,asp

# Full hybrid scan
ivta hybrid -u http://target.com -raft -e php

# Mine parameters
ivta miner -u http://target.com/api -w params.txt -m POST

# Validate reflections
ivta validator -u "http://target.com/search?q=test"

# Multiple targets
ivta fuzz -tl targets.txt -raft -e php -c 10
```

## Output

Results are saved in JSON format to the specified output file (default: `<module>_results.json`).

## Build

```bash
go build -o ivta.exe
```
