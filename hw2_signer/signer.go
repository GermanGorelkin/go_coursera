package main

import (
	"fmt"
	"sort"
	"strconv"
	_ "strconv"
	"strings"
	"sync"
	"time"
)

var wg *sync.WaitGroup
var mu *sync.Mutex

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	wg = &sync.WaitGroup{}
	mu = &sync.Mutex{}

	fmt.Println("start ExecutePipeline")

	for i, j := range jobs {
		fmt.Printf("start %v\n", i)
		wg.Add(1)
		go j(in, out)

		in = out
		out = make(chan interface{}, 100)

		//in, out = out, in

		//switch v := (<-out).(type) {
		//case int:
		//	in <- strconv.Itoa(v)
		//case string:
		//	in <- v
		//}
	}

	wg.Done()
	wg.Done()
	wg.Wait()

	time.Sleep(10 * time.Millisecond)
}

func SingleHash(in, out chan interface{}) {
	go func() {
		wgSH := &sync.WaitGroup{}

	LOOP:
		for {
			select {
			case v := <-in:
				wgSH.Add(1)
				go func() {
					dataRaw := (v).(int)
					data := strconv.Itoa(dataRaw)
					fmt.Printf("%v SingleHash data %[1]v\n", data)

					var md5Result string
					var crc32md5Result string
					var crc32Result string
					quotaCh := make(chan struct{})

					mu.Lock()
					md5Result = DataSignerMd5(data)
					fmt.Printf("%v SingleHash md5(data) %v\n", data, md5Result)
					mu.Unlock()

					go func() {
						crc32md5Result = DataSignerCrc32(md5Result)
						fmt.Printf("%v SingleHash crc32(md5(data)) %v\n", data, crc32md5Result)
						quotaCh <- struct{}{}
					}()
					go func() {
						crc32Result = DataSignerCrc32(data)
						fmt.Printf("%v SingleHash crc32(data) %v\n", data, crc32Result)
						quotaCh <- struct{}{}
					}()

					<-quotaCh
					<-quotaCh
					result := crc32Result + "~" + crc32md5Result

					fmt.Printf("%v SingleHash result %v\n", data, result)

					out <- result
					wgSH.Done()
				}()
			default:
				break LOOP
			}
		}

		wgSH.Wait()
		fmt.Println("SingleHash close")
		close(out)

		wg.Done()
	}()
}

func MultiHash(in, out chan interface{}) {
	go func() {
		wgMHInner := &sync.WaitGroup{}
		for v := range in {
			wgMHInner.Add(1)
			go func() {
				wgMH := &sync.WaitGroup{}
				dataRaw := v
				data := dataRaw.(string)

				var result string
				results := make([]string, 6, 6)
				for i := 0; i < 6; i++ {
					//crc32Result := DataSignerCrc32(string(v) + data)
					//fmt.Printf("%v MultiHash: crc32(th+step1)) %v %v\n", data, string(v), crc32Result)
					//result += crc32Result
					wgMH.Add(1)
					go func(i int) {
						th := strconv.Itoa(i)
						crc32Result := DataSignerCrc32(string(th) + data)
						fmt.Printf("%v MultiHash: crc32(th+step1)) %v %v\n", data, th, crc32Result)
						results[i] = crc32Result
						//result += crc32Result
						wgMH.Done()
					}(i)
				}
				//4108050209~502633748 MultiHash: crc32(th+step1)) 0 2956866606
				//4108050209~502633748 MultiHash: crc32(th+step1)) 0 1813119770
				wgMH.Wait()
				result = strings.Join(results, "")
				fmt.Printf("%v MultiHash result %v\n", data, result)

				out <- string(result)
				wgMHInner.Done()
			}()
		}

		wgMHInner.Wait()
		fmt.Println("MultiHash close")
		close(out)

		wg.Done()
	}()
}

func CombineResults(in, out chan interface{}) {
	go func() {
		var result []string
		for v := range in {
			dataRaw := v
			data := dataRaw.(string)

			fmt.Printf("CombineResults  %v\n", data)
			result = append(result, data)
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})

		fmt.Printf("CombineFinishResults  %v\n", strings.Join(result, "_"))

		out <- strings.Join(result, "_")
		close(out)

		wg.Done()
	}()
}
