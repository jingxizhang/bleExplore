package bleExplore

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/currantlabs/ble"
)

const (
	// PeripheralAdd    = iota
	// PeripheralRemove = iota
	retryCount = 5
)

type PeripheralAdv struct {
	Adv   ble.Advertisement
	Count int
}

var peripherals = make(map[string]*PeripheralAdv)

// Start a goroutine for discovery filtered peripheral devices
// Newly discovered perpheral is sent by action channel
func RunDiscovery(done <-chan struct{}, uuids []ble.UUID, action chan<- *PeripheralAdv) {
	// newPeripherals := make(map[string]ble.Advertisement)
	// repeatCount := 3
	uuidSet := make(map[string]struct{})
	for _, uuid := range uuids {
		// fmt.Printf("Expect Service UUID: %s\n", uuid)
		uuidSet[string(uuid)] = struct{}{}
	}

	filter := func(a ble.Advertisement) bool {
		// fmt.Printf("Filter avertisement: %s, address=%s\n", a.LocalName(), a.Address())
		for _, uuid := range a.Services() {
			// fmt.Printf("Received Service UUID: %s\n", uuid)
			if _, ok := uuidSet[string(uuid)]; ok {
				return true
			}
		}
		return false
	}

	advHandler := func(a ble.Advertisement) {
		addr := a.Address().String()
		// fmt.Printf("Received an advertisement: %s, address=%s\n",a.LocalName(), addr)
		adv, exist := peripherals[addr]
		if exist {
			adv.Count = retryCount
		} else {
			adv := &PeripheralAdv{a, retryCount}
			peripherals[addr] = adv
			action <- adv
		}
	}

	defer func() {
		fmt.Println("Discovery ended")
		close(action)
	}()

	for {
		// fmt.Println("Start scan")
		// fmt.Printf("Time = %v\n", time.Now().Second())
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 2*time.Second))
		go ble.Scan(ctx, false, advHandler, filter)
		select {
		case <-ctx.Done():
			// fmt.Printf("ctx done error type is: %v\n", ctx.Err())
			if ctx.Err() != context.DeadlineExceeded {
				return
			}
		case <-done:
			return
		}

		for k, v := range peripherals {
			// fmt.Printf("Count = %d\n", v.Count)
			if v.Count--; v.Count <= 0 {
				delete(peripherals, k)
				action <- v
			}
		}

		// fmt.Println("Sleep")
		time.Sleep(2 * time.Second)
	}
}
