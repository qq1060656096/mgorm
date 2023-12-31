package mgorm

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestConnectionConcurrent_Connect(t *testing.T) {
	config := Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}
	con := NewConnection(config)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := con.Connect()
			if err != nil {
				assert.Error(t, err, "use mysql connect error")
			}
		}()
	}
	wg.Wait()
}
