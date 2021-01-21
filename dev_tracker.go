package main

import (
    "sync"
    ws "github.com/gorilla/websocket"
)

type VidConn struct {
    socket *ws.Conn
    stopChan chan bool
}

type DevTracker struct {
    provConns map[int64] *ProviderConnection
    devToProv map[string] int64
    vidConns map[string] *VidConn 
    lock *sync.Mutex
}

func NewDevTracker() (*DevTracker) {
    self := &DevTracker{
        provConns: make( map[int64] *ProviderConnection ),
        devToProv: make( map[string] int64 ),
        lock: &sync.Mutex{},
        vidConns: make( map[string] *VidConn ),
    }
    
    return self
}

func (self *DevTracker) setVidStreamOutput( udid string, vidConn *VidConn ) {
    self.lock.Lock()
    self.vidConns[ udid ] = vidConn
    self.lock.Unlock()
}

func (self *DevTracker) getVidStreamOutput( udid string ) (*VidConn) {
    return self.vidConns[ udid ]
}

func (self *DevTracker) setDevProv( udid string, provId int64 ) {
    self.lock.Lock()
    self.devToProv[ udid ] = provId
    self.lock.Unlock()
}

func (self *DevTracker) clearDevProv( udid string ) {
    self.lock.Lock()
    delete( self.devToProv, udid )
    self.lock.Unlock()
}

func (self *DevTracker) getDevProvId( udid string ) int64 {
    return self.devToProv[ udid ]
}

func (self *DevTracker) setProvConn( provId int64, provConn *ProviderConnection ) {
    self.lock.Lock()
    self.provConns[ provId ] = provConn
    self.lock.Unlock()
}

func (self *DevTracker) getProvConn( provId int64 ) (*ProviderConnection) {
    return self.provConns[ provId ]
}

func (self *DevTracker) clearProvConn( provId int64 ) {
    self.lock.Lock()
    delete( self.provConns, provId )
    self.lock.Unlock()
}