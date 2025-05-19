package snow_flake

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// 模拟实验是生成并发400W个ID,所需要的时间
	mySnow, _ := NewSnowFlake(0, 0) // 生成雪花算法
	group := sync.WaitGroup{}
	startTime := time.Now()
	generateId := func(s *SnowFlake, requestNumber int) {
		for i := 0; i < requestNumber; i++ {
			s.NextId()
			group.Done()
		}
	}
	group.Add(4000000)
	// 生成并发的数为4000000
	currentThreadNum := 400
	for i := 0; i < currentThreadNum; i++ {
		generateId(mySnow, 10000)
	}
	group.Wait()
	fmt.Printf("time: %v\n", time.Since(startTime))
}
