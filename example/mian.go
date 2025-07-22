package main

import (
	"github.com/murang/potato/app"
	"github.com/murang/potato/net"
)

func main() {
	a := app.Instance()
	a.RegisterModule(1, &Nice{})

	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		a.SendToModule(1, "hello")
	//	}
	//}()

	// net
	netManager := net.NewManager()
	ln, err := net.NewListener("ws", "0.0.0.0:10086")
	if err != nil {
		panic(err)
	}
	netManager.AddListener(ln)
	netManager.Start()

	a.Init(nil)
	a.StartRun()
	a.Destroy(nil)
}
