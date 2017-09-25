package bleExplore

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/currantlabs/ble"
)

const (
	PeripheralAdd    = iota
	PeripheralRemote = iota
)

type PeripheralAction struct {
	Adv  ble.Advertisement
	Mode int
}

var peripherals = make(map[string]ble.Advertisement)

func RunDiscovery(done <-chan struct{}, uuids []ble.UUID, action chan<- PeripheralAction) {
	newPeripherals := []ble.Advertisement{}
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
		if _, exist := peripherals[addr]; !exist {
			peripherals[addr] = a
			newPeripherals = append(newPeripherals, a)
			action <- PeripheralAction{a, PeripheralAdd}
		}
	}

	defer func() {
		fmt.Println("Discovery ended")
		close(action)
	}()

	for {
		fmt.Println("Start scan")
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 2*time.Second))
		go ble.Scan(ctx, false, advHandler, filter)
		select {
		case <-ctx.Done():
		case <-done:
			return
		}
		fmt.Println("Sleep")
		time.Sleep(2 * time.Second)
	}

}
