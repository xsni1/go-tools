package main

type ServerPool struct {
    servers           []string
	heartBeatInterval int
	heartBeatAddr     string
}

func (sp *ServerPool) HeartBeat() {

}
