package mgorm

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrent_DbInsertOk 并发正确测试
func TestConcurrent_DbInsertOk(t *testing.T) {
	db := connectionCurrentGetDB(t)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			query := db.Where("id = ?", 1).First(&ConcurrentUser{})
			createUser := &ConcurrentUser{
				Nickname: fmt.Sprintf("demo1: TestConnectionConcurrent.DbInsertOk.%s", time.Now().Format("2006-01-02 15:04:05")),
			}
			query.Create(createUser)
		}()
	}
	wg.Wait()
}
