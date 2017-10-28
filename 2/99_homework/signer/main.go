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
	//fmt.Println(res)
	out <- res
}
func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	var md5 string
	var crc32md5 string
	var crc32 string
	for val := range in {
		item := strconv.Itoa(val.(int))
		//fmt.Println(item + " Single data " + item)
		md5 = DataSignerMd5(item)
		wg.Add(1)

		go func(item string) {
			defer wg.Done()
			crc32 = DataSignerCrc32(item)
		}(item)
		wg.Add(1)
		go func(item string) {
			defer wg.Done()
			crc32md5 = DataSignerCrc32(item)
		}(md5)
		wg.Wait()
		res := crc32 + "~" + crc32md5
		out <- res

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
	for inp := range in { //Проход по записям типа 3553453453~45353453653
		//output := make(chan int, 6)
		s := inp.(string)
		//fmt.Println(s)
		go func(s string, out chan interface{}) {
			output := make(chan int, 6)
			ch1 := make(chan string, 6)
			out1 := make(chan string, MaxInputDataLen)
			mu := &sync.Mutex{}
			mu1 := &sync.Mutex{}
			mu2 := &sync.Mutex{}

			for th := 0; th < 6; th++ { //crc хеши
				go MhtoCrc32(strconv.Itoa(th)+s, output, th, ch1)
			}
			go func(output chan int, ch chan string, out chan string, mu *sync.Mutex, mu1 *sync.Mutex) {
				//mu1 := &sync.Mutex{}
				for {
					mu.Lock()
					var counters = map[int]string{}
					for i := 0; i < 6; i++ {
						num := <-output
						counters[num] = <-ch
						//out <- counters[num]
						fmt.Println(num, counters[num])
					}
					for i := 0; i < 6; i++ {
						out <- counters[i]
					}
					mu.Unlock()
				}
			}(output, ch1, out1, mu, mu2)
			mu1.Lock()
			multRes := []string{<-out1, <-out1, <-out1, <-out1, <-out1, <-out1}
			mResultAll := strings.Join(multRes, "")
			fmt.Println(mResultAll)
			out <- mResultAll
			mu1.Unlock()
		}(s, out)
	}
} // CombineResults ...
func CombineResults(in, out chan interface{}) {
	timer := time.NewTimer(2900 * time.Millisecond)
	sum := []string{}
LOOP:
	for {
		select {
		case <-timer.C: // Timeout expired
			break LOOP
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
	time.Sleep(2900 * time.Millisecond)
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
