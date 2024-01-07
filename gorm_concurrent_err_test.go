package mgorm

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrent_DbInsertErr 并发错误测试
func TestConcurrent_DbInsertErr(t *testing.T) {
	db := connectionCurrentGetDB(t)
	query := db.Where("id = ?", 1).First(&ConcurrentUser{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			/*
				2023/12/31 21:37:58 /Users/zhaoweijie/work/develop/project/2023/golang/mgorm/connection_concurrent_test.go:110 Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'INSERT INTO `concurrent_users` (`nickname`) VALUES (?)' at line 1; Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'INSERT INTO `concurrent_users` (`nickname`) VALUES (?)' at line 1; sql: transaction has already been committed or rolled back
				[1.501ms] [rows:1] INSERT INTO `concurrent_users` (`nickname`) VALUES ('demo1: TestConnection.GetDbWithMySQL.2023-12-31 21:37:58')INSERT INTO `concurrent_users` (`nickname`) VALUES ('demo1: TestConnection.GetDbWithMySQL.2023-12-31 21:37:58')
			*/
			createUser := &ConcurrentUser{
				Nickname: fmt.Sprintf("demo1: TestConnectionConcurrent.DbInsertErr.%s", time.Now().Format("2006-01-02 15:04:05")),
			}
			query.Create(createUser)
		}()
	}
	wg.Wait()
}

// TestConcurrent_DbSelectErr 并发错误测试
func TestConcurrent_DbSelectErr(t *testing.T) {
	db := connectionCurrentGetDB(t)
	query := db.Where("id > ?", -1).Select(&ConcurrentUser{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i2 int) {
			defer wg.Done()

			/*
					2024/01/07 21:33:38 /Users/zhaoweijie/work/develop/project/2023/golang/mgorm/gorm_concurrent_err_test.go:50 Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near '?SELECT * FROM `concurrent_users`' at line 1
				[1.860ms] [rows:0] SELECT * FROM `concurrent_users` WHERE id = -1 OR id = 10 OR id = 10SELECT * FROM `concurrent_users` WHERE id = -1 OR id = 10 OR id = 10 OR id = 10 OR id = 10
				--- PASS: TestConcurrent_DbSelectErr (0.01s)
				PASS
				ok      github.com/qq1060656096/jjmgorm 0.048s
			*/
			createUser := &ConcurrentUser{
				Nickname: fmt.Sprintf("demo1: TestConnectionConcurrent.DbInsertErr.%s", time.Now().Format("2006-01-02 15:04:05")),
			}
			query.Where("id = ?", i2).Select(createUser)
		}(i)
	}
	wg.Wait()
}
