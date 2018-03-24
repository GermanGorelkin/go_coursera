package main

import (
	"fmt"
	"sync"
	"strconv"
	"strings"
	"sort"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	var in chan interface{}
	for _, val := range jobs {
		out := make(chan interface{}, 100)
		wg.Add(1)
		go func(f job, out chan interface{}, in chan interface{}) {
			f(in, out)
			close(out)
			wg.Done()
		}(val, out, in)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wgoutr := &sync.WaitGroup{}
	for num := range in {
		data := strconv.Itoa(num.(int))

		wgoutr.Add(1)
		go func(data string) {
			wg := &sync.WaitGroup{}
			var md5result string
			var crc32md5result string
			var crc32result string
			fmt.Printf("%s SingleHash data %[1]s\n", data)

			wg.Add(1)
			go func() {
				mu.Lock()
				md5result = DataSignerMd5(data)
				fmt.Printf("%s SingleHash md5(data) %s\n", data, md5result)
				mu.Unlock()
				crc32md5result = DataSignerCrc32(md5result)
				fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32md5result)
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				crc32result = DataSignerCrc32(data)
				fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32result)
				wg.Done()
			}()

			wg.Wait()
			result := crc32result + "~" + crc32md5result
			fmt.Printf("%s SingleHash result %s\n", data, result)

			out <- result
			wgoutr.Done()
		}(data)
	}

	wgoutr.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgoutr := &sync.WaitGroup{}
	for indata := range in {

		wgoutr.Add(1)
		go func() {
			wg := &sync.WaitGroup{}
			data := indata.(string)
			fmt.Printf("MultiHash data %s\n", data)
			var arr [6]string
			for i := 0; i < 6; i++ {
				par := strconv.Itoa(i) + data
				wg.Add(1)
				go func(par string, ind int) {
					iresult := DataSignerCrc32(par)
					fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", data, ind, iresult)
					arr[ind] = iresult
					wg.Done()
				}(par, i)
			}
			wg.Wait()
			result := strings.Join(arr[:], "")
			out <- result
			wgoutr.Done()
		}()
	}
	wgoutr.Wait()
}

func CombineResults(in, out chan interface{}) {
	var sl []string
	for data := range in {
		sl = append(sl, data.(string))
	}
	sort.Slice(sl, func(i, j int) bool {
		return sl[i] < sl[j]
	})
	result := strings.Join(sl, "_")
	fmt.Printf("CombineFinishResults  %v\n", result)

	out <- result
}