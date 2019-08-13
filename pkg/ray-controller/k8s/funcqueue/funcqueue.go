package funcqueue

import ()

func NoRetry(int) bool { return false }

type WaitFunc func(nRetries int) bool

type queuedFunc struct {
	f        func() error
	waitFunc WaitFunc
}

type FuncQueue struct {
	queue  chan queuedFunc
	stopCh chan struct{}
}

func NewFunctionQueue(maxLenght int32) *FuncQueue {

	fq := &FuncQueue{
		queue:  make(chan queuedFunc, maxLenght),
		stopCh: make(chan struct{}),
	}
	go fq.run()
	return fq

}

// run starts the FunctionQueue internal worker. It will be stopped once
// `stopCh` is closed or receives a value.
func (self *FuncQueue) run() {
	for {
		select {
		case <-self.stopCh:
			return
		case f := <-self.queue:
			retries := 0
			for {
				select {
				case <-self.stopCh:
					return
				default:
				}
				retries++
				if err := f.f(); err != nil {
					if !f.waitFunc(retries) {
						break
					}
				} else {
					break
				}
			}
		}
	}
}

func (self *FuncQueue) Stop() {
	close(self.stopCh)
}

func (self *FuncQueue) Enqueue(f func() error, waitFunc WaitFunc) {
	self.queue <- queuedFunc{f: f, waitFunc: waitFunc}
}
