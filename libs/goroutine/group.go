package goroutine

import "sync"

// 带并发控制的协程 Group
type Group struct {
	ch chan struct{}
	wg *sync.WaitGroup
}

// 新建 Group
// taskCount 任务大小
// goroutineCount 协程数量
func NewGroup(taskCount, goroutineCount int) (group *Group) {
	if goroutineCount > taskCount {
		goroutineCount = taskCount
	}
	wg := &sync.WaitGroup{}
	wg.Add(taskCount)
	group = &Group{
		ch: make(chan struct{}, goroutineCount),
		wg: wg,
	}
	return
}

func (this *Group) Submit(task func()) {
	this.ch <- struct{}{}
	go func() {
		defer func() {
			this.wg.Done()
			<-this.ch
		}()

		task()
	}()
}

func (this *Group) Wait() {
	this.wg.Wait()
}
