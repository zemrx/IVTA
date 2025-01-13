
---

# IVTA

A powerful and modular web crawling and fuzzing tool written in Go. This tool is designed for security testing, reconnaissance, and discovering hidden parameters, directories, and vulnerabilities in web applications.

---

## Features

- **Web Crawling**:
  - Discover links and resources on a website.
  - JavaScript rendering for crawling dynamic websites.
  - Form submission to test for vulnerabilities like SQL injection and XSS.
  - Sitemap parsing to discover additional URLs.

- **Fuzzing**:
  - Directory fuzzing to discover hidden files and directories.
  - Parameter fuzzing to discover hidden or vulnerable query parameters.
  - Reflection testing to detect reflected XSS and other vulnerabilities.

- **Hybrid Mode**:
  - Combines crawling and fuzzing into a single workflow.

- **Output**:
  - Save results to JSON.

---

## Installation

1. **Install Go**:
   - Download and install Go from [https://golang.org/dl/](https://golang.org/dl/).

2. **Clone the Repository**:
   ```bash
   git clone https://github.com/zemrx/IVTA
   cd IVTA
   ```

3. **Install Dependencies**:
   ```bash
   go get github.com/chromedp/chromedp
   go get github.com/PuerkitoBio/goquery
   go get github.com/gocolly/colly/v2
   ```

4. **Build the Tool**:
   ```bash
   go build -o ivta
   ```

---

## Usage

### Commands

The tool supports the following subcommands:

- **`crawl`**: Run the crawler.
- **`fuzz`**: Run the fuzzer.
- **`hybrid`**: Run the hybrid crawler and fuzzer.

### Examples

1. **Crawl a Website**:
   ```bash
   ./ivta crawl -u https://example.com -d 2 -c 5 -v -o crawl_results.json
   ```

2. **Fuzz for Directories and Parameters**:
   ```bash
   ./ivta fuzz -u https://example.com -w wordlist.txt -p param_wordlist.txt -c 5 -v -o fuzz_results.json -s "custom_value"
   ```

3. **Run Hybrid Crawling and Fuzzing**:
   ```bash
   ./ivta hybrid -u https://example.com -w wordlist.txt -p param_wordlist.txt -d 2 -c 5 -v -o hybrid_results.json -s "custom_value"
   ```

### Options

- **`-u`**: Target URL (required).
- **`-w`**: Path to the directory wordlist file (default: `wordlist.txt`).
- **`-p`**: Path to the parameter wordlist file (default: `param_wordlist.txt`).
- **`-d`**: Maximum depth for recursive discovery (default: `2`).
- **`-c`**: Number of concurrent requests (default: `5`).
- **`-v`**: Enable verbose mode.
- **`-o`**: Path to the output file (default: `results.json`).
- **`-s`**: Custom symbol to test for reflection (default: `test`).
- **`-h`**: Display help message.

---

## Wordlists

The tool uses wordlists for directory and parameter fuzzing. You can use your own wordlists or download popular ones like:

- [SecLists](https://github.com/danielmiessler/SecLists)
- [FuzzDB](https://github.com/fuzzdb-project/fuzzdb)

Place your wordlists in the `wordlists/` directory or specify the path using the `-w` and `-p` options.

---

## Configuration

The tool supports configuration via command-line arguments. For advanced configurations, you can modify the source code or use environment variables.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contact

For questions or feedback, please open an issue or contact [zemrx0@gmail.com](mailto:zemrx0@gmail.com).

---
