package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for val := range in {
		var md5 string
		var crc32md5 string
		var crc32 string
		item := strconv.Itoa(val.(int))
		md5 = DataSignerMd5(item)
		wg1 := &sync.WaitGroup{}
		wg.Add(1)
		go func(item string, wg1 *sync.WaitGroup) {
			defer wg.Done()
			wg1.Add(1)
			go func(item string) {
				defer wg1.Done()
				crc32 = DataSignerCrc32(item)
			}(item)
			wg1.Add(1)
			go func(item string) {
				defer wg1.Done()
				crc32md5 = DataSignerCrc32(item)
			}(md5)
			wg1.Wait()
			res := crc32 + "~" + crc32md5
			out <- res
			//wg1.Wait()
		}(item, wg1)
		//wg1.Wait()
	}
	wg.Wait()
}

func Res(in string) string {
	var wg sync.WaitGroup
	st := make([]string, 6)
	for index := range st {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st[i] = DataSignerCrc32(strconv.Itoa(i) + in)
			// fmt.Printf("%v MultiHash: crc32(th+step1)) %v %v\n", in, i, st[i])
		}(index)
	}
	wg.Wait()
	out := strings.Join(st, "")
	// fmt.Printf("%v MultiHash result:\n %v\n", in, out)
	return out
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for data := range in {
		//fmt.Println(data)

		wg.Add(1)

		go func(data string) {
			defer wg.Done()
			out <- Res(data)
		}(data.(string))
	}
	wg.Wait()
}
func CombineResults(in, out chan interface{}) {
	//timer := time.NewTimer(2900 * time.Millisecond)
	sum := []string{}
LOOP:
	for {
		select {
		//	case <-timer.C: // Timeout expired
		//		break LOOP
		case res, ok := <-in:
			//fmt.Println(res)
			if !ok { // Channel closed
				break LOOP
			}
			val, ok := res.(string)
			if ok { // Channel closed
				sum = append(sum, val)
			}
		}
	}
	sort.Strings(sum)
	//fmt.Println(sum)
	out <- strings.Join(sum, "_")
	//close(out)
}

//inputData := []int{0, 1}
func ExecutePipeline(hSP ...job) {
	var wg sync.WaitGroup
	runtime.GOMAXPROCS(0)
	in := make(chan interface{}, 10)
	for _, work := range hSP {
		wg.Add(1)
		out := make(chan interface{}, 10)
		go func(in, out chan interface{}, work job) {
			defer wg.Done()
			work(in, out)
			close(out)
		}(in, out, work)
		in = out
	}
	//time.Sleep(2900 * time.Millisecond)
	wg.Wait()
}
func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	testResult := "NOT_SET"
	testExpected := "1173136728138862632818075107442090076184424490584241521304_1696913515191343735512658979631549563179965036907783101867_27225454331033649287118297354036464389062965355426795162684_29568666068035183841425683795340791879727309630931025356555_3994492081516972096677631278379039212655368881548151736_4958044192186797981418233587017209679042592862002427381542_4958044192186797981418233587017209679042592862002427381542"

	hashSignPipeline := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
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
	if testResult == testExpected {
		fmt.Println("good")
	} else if testResult != testExpected {
		fmt.Println("not good")
	}
}
