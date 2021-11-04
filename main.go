package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {
	word := os.Args[1]

	perms := make([]string,0)

	// common names that are generally present in a bucket name
	envs := []string{"dev","development","stage","stg","s3","staging","prod","production","test"}

	// appending the company name itself
	perms = append(perms, word)

	// allowed special chars according to aws bucket naming rules
	ds := []string{".","-",""} 

	// common_bucket_prefixes.txt contains a list of most common prefixes in a bucket name (of course😅), thanks to @nahamsec
	file, err := os.Open("common_bucket_prefixes.txt")
	if err != nil {
		fmt.Println("File not found!")
	} 
	list, err := ioutil.ReadAll(file)
	prefix_list := strings.Split(strings.ReplaceAll(string(list),"\r\n","\n"),"\n")

	// adding (1*(3)*9)*2 = 54 permutations (e.g. uber.dev, stg-uber, s3uber i.e. words from envs)
	for _, env := range envs {
		for _, i := range ds {
			perms = append(perms, word+i+env)
			perms = append(perms, env+i+word)
		}
	}

	// adding (1*(3)*198)*2 = 1188 permutations (e.g. admin-uber, uberbucket, backup.uber i.e. from common_bucket_prefixes.txt)
	for _, env := range prefix_list {
		for _, i := range ds {
			perms = append(perms, word+i+env)
			perms = append(perms, env+i+word)
		}
	}

	// adding (1*(3)*198*(3)*9)*6 = 96228 permutations
	for _, i := range ds {
		for _, prefix := range prefix_list {
			for _, j := range ds {
				for _, env := range envs {
					perms = append(perms, word+i+prefix+j+env)
					perms = append(perms, word+i+env+j+prefix)
					perms = append(perms, prefix+i+word+j+env)
					perms = append(perms, env+i+word+j+prefix)
					perms = append(perms, prefix+i+env+j+word)
					perms = append(perms, env+i+prefix+j+word)
				}
			}
		}
	}

	// removing duplicate permutations and all the words that have length > 63 (not allowed in bucket naming rules)
	check := make(map[string]int)
	for _, i := range perms {
		if len(i) <= 63 {
			check[i] = 1
		}
	}
	res := make([]string,0)
	for str := range check {
		res = append(res, str)
	}

	// colorRed := "\033[31m"
    // colorGreen := "\033[32m"

	jobs := make(chan string)
	var wg sync.WaitGroup

	for k := 0; k < 50; k++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for prefix := range jobs {
				resp, err := http.Get("http://"+prefix+".s3.amazonaws.com")
				if err != nil {
					continue
				}
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					
					fmt.Printf("http://%s.s3.amazonaws.com: %d\n", prefix, resp.StatusCode)
				} else if resp.StatusCode < 404 {
					fmt.Printf("http://%s.s3.amazonaws.com: %d\n", prefix, resp.StatusCode)
				}
			}
		}()
	}
	for _, prefix := range res {
		jobs <- prefix
	}
	close(jobs)

	wg.Wait()
}