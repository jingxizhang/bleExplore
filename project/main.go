package main

import (
	"fmt"
	"log"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/examples/lib/dev"
	"github.com/jingxizhang/bleExplore"
)

func main() {
	done := make(chan struct{})
	action := make(chan *bleExplore.PeripheralAdv)

	filterUuids := []ble.UUID{
		ble.MustParse("71A0"),
		ble.MustParse("71A2"),
		ble.MustParse("71A3"),
	}

	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	fmt.Println("Start run discovery")
	go bleExplore.RunDiscovery(done, filterUuids, action)

	defer close(done)

	for act := range action {
		if act.Count > 0 {
			fmt.Printf("Add a peripheral: %s at %v\n",
				act.Adv.LocalName(), act.Adv.Address().String())
		} else {
			fmt.Printf("Remove a peripheral: %s at %v\n",
				act.Adv.LocalName(), act.Adv.Address().String())
		}
	}

	fmt.Println("Program Ended")
}
