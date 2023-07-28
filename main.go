package main

func main() {
	runProxy(config{Connection: struct {
		LocalAddress  string
		RemoteAddress string
	}{LocalAddress: "0.0.0.0:19132", RemoteAddress: "zeqa.net:19132"}, AuthEnabled: true})
	//runServer()
}
