package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"bufio"
	"os"
	"net"
	"net/http"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	"flag"
	"fmt"
	"time"
)



func Resolver(domain string) string {
	var (
		dnsResolverIP        = "8.8.8.8:53" 
		dnsResolverProto     = "udp"        
		dnsResolverTimeoutMs = 5000        
	)

	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
	httpClient := &http.Client{}
	resp, err := httpClient.Get(fmt.Sprintf("https://%s", domain))
	if err != nil {
		return "false"
	}
	defer resp.Body.Close()

	return "true"
}

func subfinderSingle(domain string) string {
	subfinderOpts := &runner.Options{
		Threads:            10,
		Timeout:            30,
		MaxEnumerationTime: 10,
		All:                true,
		ProviderConfig:     "./providers.yaml",
	}
	log.SetFlags(0)

	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		log.Fatalf("failed to create subfinder runner: %v", err)
	}
	output := &bytes.Buffer{}
	if err = subfinder.EnumerateSingleDomainWithCtx(context.Background(), domain, []io.Writer{output}); err != nil {
		log.Fatalf("failed to enumerate single domain: %v", err)
	}
	return output.String()
}


func subfinderRootFile(filename string) string {
	subfinderOpts := &runner.Options{
		Threads:            10,
		Timeout:            30,
		MaxEnumerationTime: 10,
		All:                true,
		ProviderConfig:     "./providers.yaml",
	}
	log.SetFlags(0)
	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		log.Fatalf("failed to create subfinder runner: %v", err)
	}

	output := &bytes.Buffer{}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open domains file: %v", err)
	}
	defer file.Close()

	if err = subfinder.EnumerateMultipleDomainsWithCtx(context.Background(), file, []io.Writer{output}); err != nil {
		log.Fatalf("failed to enumerate subdomains from file: %v", err)
	}
	return output.String()
}
func compareAndGetUniqueLines(file1, file2 string) ([]string, error) {
	lines1, err := readLines(file1)
	if err != nil {
		return nil, err
	}

	lines2, err := readLines(file2)
	if err != nil {
		return nil, err
	}

	uniqueLines := []string{}
	for _, line := range lines2 {
		if !contains(lines1, line) {
			uniqueLines = append(uniqueLines, line)
		}
	}

	return uniqueLines, nil
}

func contains(lines []string, line string) bool {
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false

}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func saveFile(filename string, data string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString(data)
}

func Webhook(data string) {
	Discordwebhook := "https://discord.com/api/webhooks/1200485943946264606/1s42QMwqMbDQWf1FCIxKkxFBJ6VXPd2v-_l1rQRjSDO5Jjxbv3ib_LITM77RiuDIndZb"
	jsonStr := []byte(`{"content":"` + data + `"}`)
	req, err := http.NewRequest("POST", Discordwebhook, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

}

func ManageEngineer(option string, domain string, filename string) {
	if option == "single" {
		data := subfinderSingle(domain)
		oldfileName := domain + "_main_subdomains.txt"
		if _, err := os.Stat(oldfileName); err == nil {
			newFilename := domain + "_new_subdomains.txt"
			saveFile(newFilename, data)
			uniqueLines, err := compareAndGetUniqueLines(oldfileName, newFilename)
			if err != nil {
				log.Fatal(err)
			}
			for _, line := range uniqueLines {
				resolve := Resolver(line)
				if resolve == "true" {
					Webhook(line+" **:Live**")
				} else {
					Webhook(line+" **:Dead**")

				}
			}
			saveFile(oldfileName, data)
		} else {
			saveFile(oldfileName, data)
		}
	} else if option == "file" {
		data := subfinderRootFile(filename)
		oldfileName := filename + "_main_subdomains.txt"
		if _, err := os.Stat(oldfileName); err == nil {
			newFilename := filename + "_new_subdomains.txt"
			saveFile(newFilename, data)
			uniqueLines, err := compareAndGetUniqueLines(oldfileName, newFilename)
			if err != nil {
				log.Fatal(err)
			}
			for _, line := range uniqueLines {
				resolve := Resolver(line)
				if resolve == "true" {
					Webhook(line+":Live")
				} else {
					Webhook(line+":Dead")

				} 
			}
			saveFile(oldfileName, data)
		} else {
			saveFile(oldfileName, data)
		}
	} else {
		log.Fatal("Invalid Option")
	}
}

func TimeTable24Hours(option string, domain string, filename string) {
	nextRun := time.Now().Add(24 * time.Hour)
	if "single" == option {
		for {
			ManageEngineer(option, domain, filename)
			fmt.Println("[INF] Next run at", nextRun.Format("2006-01-02 15:04:05"))
			time.Sleep(24 * time.Hour)
		}
	} else if "file" == option {
		for {
			ManageEngineer(option, domain, filename)
			fmt.Println("[INF] Next run at", nextRun.Format("2006-01-02 15:04:05"))
			time.Sleep(24 * time.Hour)
			
		}
	} else {
		log.Fatal("Invalid Option")
	}
}





func main() {
	fmt.Println(`
 ___      _      _   _         _   
/ __|_  _| |__  /_\ | |___ _ _| |_ 
\__ \ || | '_ \/ _ \| / -_) '_|  _|
|___/\_,_|_.__/_/ \_\_\___|_|  \__|
--Project Subalert
	`)
	var domain string
	var filename string
	flag.StringVar(&domain, "d", "", "Domain to enumerate")
	flag.StringVar(&filename, "f", "", "File to enumerate")
	flag.Parse()
	if domain != "" {
		TimeTable24Hours("single", domain, filename)
	}
	if filename != "" {
		TimeTable24Hours("file", domain, filename)
	}



}
