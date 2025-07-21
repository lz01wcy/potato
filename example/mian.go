package main

import (
	"github.com/murang/potato/app"
	"time"
)

func main() {
	a := app.Instance()
	a.RegisterModule(1, &Nice{})

	go func() {
		for {
			time.Sleep(time.Second)
			a.SendToModule(1, "hello")
		}
	}()

	a.Init(nil)
	a.StartRun()
	a.Destroy(nil)
}
