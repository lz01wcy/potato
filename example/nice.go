package main

type Nice struct {
}

func (n Nice) FPS() uint {
	return 3
}

func (n Nice) Start() {
	println("start")
}

func (n Nice) Update() {
	println("update")
}

func (n Nice) OnDestroy() {
	println("destroy")
}

func (n Nice) OnMsg(msg interface{}) {
	println("msg: " + msg.(string))
}

func (n Nice) OnRequest(msg interface{}) interface{} {
	println("request")
	return nil
}
