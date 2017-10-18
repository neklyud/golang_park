package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
)

//func md5Calc(string data,chan interface {} in) chan out{
//	in<-DataSignerMd5(data)
//}
func calcmd5(data string, md5 chan string, cancCh chan struct{}) {
	res := DataSignerMd5(data)
	md5 <- res
	//fmt.Println(data + " Signer md5(data) " + res)
}

func calccrc32(data string, crc32 chan string, item string) {
	res := DataSignerCrc32(data)
	crc32 <- res
	//fmt.Println(item + " Signer crc32(data) " + res)
}

func calccrc32md5(data chan string, crc32 chan string, item string) {
	datastring := <-data
	res := DataSignerCrc32(datastring)
	crc32 <- res
	//fmt.Println(item + " Signer crc32(md5(data)) " + res)
}
func concat(one chan string, two chan string, out chan interface{}) {
	first := <-one
	second := <-two
	res := first + "~" + second
	fmt.Println(res)
	out <- res
}
func SingleHash(in, out chan interface{}) {
	md5 := make(chan string, MaxInputDataLen)
	crc32 := make(chan string, MaxInputDataLen)
	crcmd5 := make(chan string, MaxInputDataLen)
	cancCh := make(chan struct{}, 1)
	for {
		item := strconv.Itoa((<-in).(int))
		//fmt.Println(item + " Single data " + item)
		calcmd5(item, md5, cancCh)
		go calccrc32(item, crc32, item)
		go calccrc32md5(md5, crcmd5, item)
		go concat(crc32, crcmd5, out)
		//c := <-out
		//fmt.Println(c)
	}
}
func MhtoCrc32(str string, out chan int, th int, in chan string) {
	res := DataSignerCrc32(str)
	in <- res
	out <- th
}

func MultiHash(in, out chan interface{}) {
	ch1 := make(chan string, MaxInputDataLen)
	for inp := range in {
		output := make(chan int, 5)
		s := inp.(string)
		//fmt.Println(s)
		go func(s string) {
			cancelCh := make(chan struct{})
			mu := &sync.Mutex{}
			for th := 0; th < 5; th++ {
				go MhtoCrc32(strconv.Itoa(th)+s, output, th, ch1)
			}
			go func(output chan int, ch chan string, out chan interface{}, cancelCh chan struct{}, mu *sync.Mutex) {
				var counters = map[int]string{}
				for {
					mu.Lock()
					for i := 0; i < 5; i++ {
						num := <-output
						counters[num] = <-ch
					}
					for i := 0; i < 5; i++ {
						fmt.Println(i, counters[i])
						//out <- counters[i]
					}
					//out <- "_"
					mu.Unlock()
				}
			}(output, ch1, out, cancelCh, mu)
			//cancelCh <- struct{}{}
		}(s)
	}
}

//inputData := []int{0, 1}
func ExecutePipeline(hSP ...job) {
	runtime.GOMAXPROCS(0)
	in := make(chan interface{}, 1)
	for _, work := range hSP {
		out := make(chan interface{}, 1)
		go work(in, out)
		in = out
	}
	time.Sleep(10000 * time.Millisecond)
}
func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	testResult := "NOT_SET"

	hashSignPipeline := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		//job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			//fmt.Println(123)
			<-in
			data, _ := dataRaw.(string)
			testResult = data
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignPipeline...)

	end := time.Since(start)

	fmt.Println(end)
	fmt.Println(testResult)
}
