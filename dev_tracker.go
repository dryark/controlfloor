package main

import (
	ws "github.com/gorilla/websocket"
	"sync"
)

type VidConn struct {
	socket   *ws.Conn
	stopChan chan bool
	offset   int64
}

type DevStatus struct {
	wda   bool
	cfa   bool
	video bool
}

type DevTracker struct {
	provConns map[int64]*ProviderConnection
	devToProv map[string]int64
	DevStatus map[string]*DevStatus
	vidConns  map[string]*VidConn
	clients   map[string]chan ClientMsg
	lock      *sync.Mutex
	config    *Config
}

func NewDevTracker(config *Config) *DevTracker {
	self := &DevTracker{
		provConns: make(map[int64]*ProviderConnection),
		devToProv: make(map[string]int64),
		lock:      &sync.Mutex{},
		vidConns:  make(map[string]*VidConn),
		DevStatus: make(map[string]*DevStatus),
		clients:   make(map[string]chan ClientMsg),
		config:    config,
	}

	return self
}

func (self *DevTracker) setVidStreamOutput(udid string, vidConn *VidConn) {
	self.lock.Lock()
	self.vidConns[udid] = vidConn
	self.lock.Unlock()
}

func (self *DevTracker) getVidStreamOutput(udid string) *VidConn {
	return self.vidConns[udid]
}

func (self *DevTracker) setDevProv(udid string, provId int64) {
	self.lock.Lock()
	self.devToProv[udid] = provId
	self.DevStatus[udid] = &DevStatus{}
	self.lock.Unlock()
}

func (self *DevTracker) clearDevProv(udid string) {
	self.lock.Lock()
	delete(self.devToProv, udid)
	delete(self.DevStatus, udid)
	self.lock.Unlock()
}

func (self *DevTracker) addClient(udid string, msgChan chan ClientMsg) {
	self.lock.Lock()
	self.clients[udid] = msgChan
	self.lock.Unlock()
}

func (self *DevTracker) deleteClient(udid string) {
	self.lock.Lock()
	delete(self.clients, udid)
	self.lock.Unlock()
}

func (self *DevTracker) msgClient(udid string, msg ClientMsg) {
	msgChan, chanOk := self.clients[udid]
	if !chanOk {
		return
	}
	msgChan <- msg
}

func (self *DevTracker) setDevStatus(udid string, service string, status bool) {
	stat, statOk := self.DevStatus[udid]
	if !statOk {
		return
	}
	if service == "wda" {
		stat.wda = status
		return
	}
	if service == "cfa" {
		stat.cfa = status
		return
	}
	if service == "video" {
		stat.video = status
		return
	}
}

func (self *DevTracker) getDevStatus(udid string) *DevStatus {
	devStatus, devOk := self.DevStatus[udid]
	if devOk {
		return devStatus
	} else {
		return nil
	}
}

func (self *DevTracker) getDevProvId(udid string) int64 {
	provId, provOk := self.devToProv[udid]
	if provOk {
		return provId
	} else {
		return 0
	}
}

func (self *DevTracker) setProvConn(provId int64, provConn *ProviderConnection) {
	self.lock.Lock()
	self.provConns[provId] = provConn
	self.lock.Unlock()
}

func (self *DevTracker) getProvConn(provId int64) *ProviderConnection {
	return self.provConns[provId]
}

func (self *DevTracker) clearProvConn(provId int64) {
	self.lock.Lock()
	delete(self.provConns, provId)
	self.lock.Unlock()
}
