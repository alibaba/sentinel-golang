package etcdv3

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type Watch struct {
	revision int64
	client   *clientv3.Client
	cancel   context.CancelFunc

	eventChan chan *clientv3.Event
	key string
}

// C ...
func (w *Watch) C() chan *clientv3.Event {
	return w.eventChan
}

func (w *Watch) Close() {
	w.cancel()
}

func (w *Watch) update(resp *clientv3.WatchResponse) {
	if resp.CompactRevision > w.revision {
		w.revision = resp.CompactRevision
	} else if resp.Header.GetRevision() > w.revision {
		w.revision = resp.Header.GetRevision()
	}

	if err := resp.Err(); err != nil {
		log.Printf("resp err = %+v\n", err)
		return
	}

	for _, event := range resp.Events {
		select {
		case w.eventChan <- event:
		default:
			log.Println("blocked event chan")
		}
	}
}

// NewWatch ...
func newWatch(client *clientv3.Client, key string) *Watch {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		watcher     = &Watch{
			client:    client,
			revision:  0,
			cancel:    cancel,
			eventChan: make(chan *clientv3.Event, 100),
			key: key,
		}
	)

	go func() {
		rch := client.Watch(ctx, key, clientv3.WithCreatedNotify())
		for {
			for resp := range rch {
				watcher.update(&resp)
			}

			time.Sleep(time.Duration(1) * time.Second)
			if watcher.revision > 0 {
				rch = client.Watch(ctx, key, clientv3.WithCreatedNotify(), clientv3.WithRev(watcher.revision))
			} else {
				rch = client.Watch(ctx, key, clientv3.WithCreatedNotify())
			}
		}
	}()

	return watcher
}

