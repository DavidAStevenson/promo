package main

import (
	"fmt"
	"time"

	"promo"
)

func main() {
	now := time.Now()
	start := now.Add(2 * time.Second)
	end := now.Add(5 * time.Second)
	b1, err1 := promo.New("Demo banner", start, end)
	if err1 != nil {
		fmt.Printf("Created promo: %v, status: %v\n", b1.Name(), b1.Status())
	}

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		b := promo.Choose("192.168.0.1")
		if b == nil {
			fmt.Printf("%d: No promo available now\n", i)
		} else {
			fmt.Printf("%d: Selected promo: %v, status is %v\n", i, b.Name(), b.Status())
		}
	}
}
