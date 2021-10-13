package common

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	PREFIX_XX = "012345678901234567890123456789012345678901234567890123456789_"
)

func LevelDb() {
	op := opt.Options{
		// WriteBuffer: 40 * opt.MiB,
		// Compression: opt.DefaultCompressionType,
		// CompactionTableSize:           100 * opt.MiB,
		// CompactionTableSizeMultiplier: 10,
	}
	db, err := leveldb.OpenFile(filepath.Join(DATA_DIR, "leveldb"), &op)
	if err != nil {
		log.Fatal(err)
	}

	var cnt int64 = 0
	var sum int64 = 0

	// print DBStats
	go func() {
		var stat leveldb.DBStats
		for _ = range time.Tick(time.Second * 5) {
			err := db.Stats(&stat)
			if err != nil {
				log.Print(err)
			}
			if stat.WritePaused {
				log.Printf("%+v", stat)
			}
			fmt.Println("Db has records: ", atomic.LoadInt64(&sum))

			if stat.WritePaused && stat.LevelTablesCounts[0] < opt.DefaultWriteL0PauseTrigger {
				log.Printf("****************** still in WritePaused **************")
			}

		}
	}()

	wg := sync.WaitGroup{}

	// scan
	START := time.Now()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			l := list.New()
			start := time.Now()
			var num int64
			it := db.NewIterator(&util.Range{}, nil)
			for it.Next() {
				l.PushBack(string(it.Key()))
				_ = it.Value()
				num++
				atomic.StoreInt64(&cnt, num)
			}
			spt := time.Since(start)
			qps := time.Since(START)
			if qps.Milliseconds() != 0 {
				fmt.Println("Start spend ", qps.Seconds(), "s, ", spt.Milliseconds(), "ms, qps: ", num*1000/qps.Milliseconds())
			}
			it.Release()

			fmt.Println("Total records: ", cnt)
			for e := l.Front(); e != nil; e = e.Next() {
				db.Delete([]byte(e.Value.(string)), nil)
			}
			qps = time.Since(START)
			if qps.Milliseconds() != 0 {
				fmt.Println("Start spend ", qps.Seconds(), "s, ", spt.Milliseconds(), "ms, qps: ", num*1000/qps.Milliseconds())
			}

			if num > 10000000 {
				os.Exit(-1)
			}
		}

	}()

	// var value = make([]byte, 20)
	// for i := 0; i < 100; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		// write in batch
	// 		defer wg.Done()
	// 		var batch leveldb.Batch
	// 		for {
	// 			start := time.Now()
	// 			batch.Reset()
	// 			idx := rand.Int()
	// 			for i := idx; i < idx+1; i++ {
	// 				batch.Put([]byte(PREFIX_XX+strconv.Itoa(i)), value)
	// 				atomic.AddInt64(&sum, 1)
	// 			}
	// 			db.Write(&batch, nil)

	// 			if atomic.LoadInt64(&sum) > 10000000 {
	// 				return
	// 			}

	// 			spt := time.Since(start)
	// 			if spt.Seconds() > 1 {
	// 				fmt.Println("Write spend ", spt.Milliseconds(), "ms.")
	// 			}
	// 		}
	// 	}()
	// }

	// for i := 0; i < 100; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		// delete in batch
	// 		defer wg.Done()
	// 		var batch leveldb.Batch
	// 		for {
	// 			start := time.Now()
	// 			batch.Reset()
	// 			idx := rand.Int()
	// 			for i := idx; i < idx+1; i++ {
	// 				batch.Delete([]byte(PREFIX_XX + strconv.Itoa(i)))
	// 			}
	// 			db.Write(&batch, nil)
	// 			spt := time.Since(start)
	// 			if spt.Seconds() > 1 {
	// 				fmt.Println("Delete spend ", spt.Milliseconds(), "ms.")
	// 			}
	// 		}
	// 	}()
	// }

	wg.Wait()
}
