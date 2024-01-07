package mgorm

import (
	"fmt"
	"gorm.io/gorm"
	"sync"
	"testing"
)

func TestConcurrentErr(t *testing.T) {
	db := connectionCurrentGetDB(t)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int, db *gorm.DB) {
			defer wg.Done()
			db.Create(&ConcurrentUser{Nickname: fmt.Sprintf("nickname-%d", i)})
		}(i, db)
	}
	wg.Wait()
}
