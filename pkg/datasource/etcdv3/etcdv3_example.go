package etcdv3

import (
	"fmt"
	"log"
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/mock"
)

// New one datasource based on etcv3 client
func Example_ClientWithOneDatasource() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cli.Close()

	h := datasource.MockPropertyHandler{}
	h.On("isPropertyConsistent", mock.Anything).Return(true)
	h.On("Handle", mock.Anything).Return(nil)
	ds, err := NewDataSource(cli, "foo", &h)
	if err != nil {
		log.Fatal(err)
	}
	err = ds.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(30 * time.Second)
	fmt.Println("Prepare to close")
	ds.Close()
	time.Sleep(120 * time.Second)
}

// New multi datasource based on etcv3 client
func Example_ClientWithMultiDatasource() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal("client error:", err)
	}
	defer cli.Close()

	h := datasource.MockPropertyHandler{}
	h.On("isPropertyConsistent", mock.Anything).Return(true)
	h.On("Handle", mock.Anything).Return(nil)
	ds1, err := NewDataSource(cli, "foo", &h)
	if err != nil {
		log.Fatal(err)
	}

	ds2, err := NewDataSource(cli, "aoo", &h)
	if err != nil {
		log.Fatal(err)
	}

	err = ds1.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	err = ds2.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(60 * time.Second)
	fmt.Println("Prepare to close ds1")
	// close ds1, will not recv the watch response
	ds1.Close()

	// ds2 also could recv the watch response
	time.Sleep(80 * time.Second)
	ds2.Close()

	time.Sleep(100 * time.Second)
}
