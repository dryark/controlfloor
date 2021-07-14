package main

import (
    "fmt"
    uj "github.com/nanoscopic/ujsonin/v2/mod"
)

type ProvBase interface {  
    asText( int16 ) string
    needsResponse() bool
    resHandler() (func(uj.JNode))
}

type ProvPing struct {
    blah string
    onRes func( uj.JNode )
}
func (self *ProvPing) resHandler() (func(uj.JNode) ) { return self.onRes }
func (self *ProvPing) needsResponse() (bool) { return true }
func (self *ProvPing) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"ping\"}\n", id)
}

type ProvClick struct {
    udid string
    x int
    y int
    onRes func( uj.JNode )
}
func (self *ProvClick) resHandler() ( func(data uj.JNode) ) {
    return self.onRes
}
func (self *ProvClick) needsResponse() (bool) { return true }
func (self *ProvClick) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"click\",udid:\"%s\",x:%d,y:%d}\n",id,self.udid,self.x,self.y)
}

type ProvHardPress struct {
    udid string
    x int
    y int
}
func (self *ProvHardPress) resHandler() (func(uj.JNode) ) { return nil }
func (self *ProvHardPress) needsResponse() (bool) { return false }
func (self *ProvHardPress) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"hardPress\",udid:\"%s\",x:%d,y:%d}\n",id,self.udid,self.x,self.y)
}

type ProvLongPress struct {
    udid string
    x int
    y int
}
func (self *ProvLongPress) resHandler() (func(uj.JNode) ) { return nil }
func (self *ProvLongPress) needsResponse() (bool) { return false }
func (self *ProvLongPress) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"longPress\",udid:\"%s\",x:%d,y:%d}\n",id,self.udid,self.x,self.y)
}

type ProvHome struct {
    udid string
    onRes func( uj.JNode )
}
func (self *ProvHome) resHandler() ( func(data uj.JNode) ) {
    return self.onRes
}
func (self *ProvHome) needsResponse() (bool) { return true }
func (self *ProvHome) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"home\",udid:\"%s\"}\n",id,self.udid)
}

type ProvKeys struct {
    udid string
    keys string
    curid int
    prevkeys string
    onRes func( uj.JNode )
}
func (self *ProvKeys) resHandler() ( func(data uj.JNode) ) {
    return self.onRes
}
func (self *ProvKeys) needsResponse() (bool) { return true }
func (self *ProvKeys) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"keys\",udid:\"%s\",keys:\"%s\",curid:%d,prevkeys:\"%s\"}\n",
      id,self.udid,self.keys,self.curid,self.prevkeys)
}

type ProvSwipe struct {
    udid string
    x1 int
    y1 int
    x2 int
    y2 int
    delay float64
    onRes func( uj.JNode )
}
func (self *ProvSwipe) resHandler() ( func(data uj.JNode) ) {
    return self.onRes
}
func (self *ProvSwipe) needsResponse() (bool) { return true }
func (self *ProvSwipe) asText( id int16 ) (string) {
    delayBy100 := int( self.delay * 100 )
    return fmt.Sprintf("{id:%d,type:\"swipe\",udid:\"%s\",x1:%d,y1:%d,x2:%d,y2:%d,delay:%d}\n",
        id,self.udid,self.x1,self.y1,self.x2,self.y2,delayBy100)
}

type ProvStartStream struct {
    udid string
}
func (self *ProvStartStream) resHandler() (func(uj.JNode) ) { return nil }
func (self *ProvStartStream) needsResponse() (bool) { return false }
func (self *ProvStartStream) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"startStream\",udid:\"%s\"}\n",id,self.udid)
}

type ProvStopStream struct {
    udid string
}

func (self *ProvStopStream) resHandler() (func(uj.JNode) ) {
    return nil
}

func (self *ProvStopStream) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"stopStream\",udid:\"%s\"}\n",id,self.udid)
}

func (self *ProvStopStream) needsResponse() (bool) {
    return false
}