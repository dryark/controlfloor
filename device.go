package main

import (
    "fmt"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
)

type DevHandler struct {
    providerAuthGroup *gin.RouterGroup
    userAuthGroup     *gin.RouterGroup
    devTracker        *DevTracker
    sessionManager    *cfSessionManager
}

func NewDevHandler(
    providerAuthGroup *gin.RouterGroup,
    userAuthGroup     *gin.RouterGroup,
    devTracker        *DevTracker,
    sessionManager    *cfSessionManager,
) *DevHandler {
    return &DevHandler{
        providerAuthGroup,
        userAuthGroup,
        devTracker,
        sessionManager,
    }
}

func (self *DevHandler) registerDeviceRoutes() {
    pAuth := self.providerAuthGroup
    uAuth := self.userAuthGroup
    
    fmt.Println("Registering device routes")
    //pAuth.GET("/devStatus", showDevStatus )
    pAuth.POST("/devStatus", func( c *gin.Context ) {
        self.handleDevStatus( c )
    } )
    // - Device is present on provider
    // - Device Info fetched from device
    // - WDA start/stop
    // - Video Streamer start/stop
    // - Video seems active/inactive
    
    //uAuth.GET("/devClick", showDevClick )
    uAuth.POST("/devClick",     func( c *gin.Context ) { self.handleDevClick( c ) } )
    uAuth.POST("/devHardPress", func( c *gin.Context ) { self.handleDevHardPress( c ) } )
    uAuth.POST("/devLongPress", func( c *gin.Context ) { self.handleDevLongPress( c ) } )
    uAuth.POST("/devHome",      func( c *gin.Context ) { self.handleDevHome( c ) } )
    uAuth.POST("/devSwipe",     func( c *gin.Context ) { self.handleDevSwipe( c ) } )
    uAuth.POST("/keys",         func( c *gin.Context ) { self.handleKeys( c ) } )
    
    uAuth.GET("/devInfo", func( c *gin.Context ) {
        self.showDevInfo( c )
    } )
    
    uAuth.GET("/devVideo", self.showDevVideo )
    
    uAuth.GET("/devPing", self.handleDevPing )
}

func (self *DevHandler) showDevInfo( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "devInfo", gin.H{
            "udid": "?",
            "name": "?",
            "clickWidth": "?",
            "clickHeight": "?",
        } )
        return
    }
    
    dev := getDevice( udid )
    if dev == nil {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no dev with that udid",
        } )
        return
    }
    
    c.HTML( http.StatusOK, "devInfo", gin.H{
        "udid": udid,
        "name": dev.Name,
        "clickWidth": dev.ClickWidth,
        "clickHeight": dev.ClickHeight,
    } )
}

func (self *DevHandler) getPc( c *gin.Context ) (*ProviderConnection,string) {
    udid := c.PostForm("udid")
    provId := self.devTracker.getDevProvId( udid )
    provConn := self.devTracker.getProvConn( provId )
    return provConn, udid
}

func (self *DevHandler) handleDevClick( c *gin.Context ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := self.getPc( c )
    pc.doClick( udid, x, y )
}

func (self *DevHandler) handleDevHardPress( c *gin.Context ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := self.getPc( c )
    pc.doHardPress( udid, x, y )
}

func (self *DevHandler) handleDevLongPress( c *gin.Context ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := self.getPc( c )
    pc.doLongPress( udid, x, y )
}

func (self *DevHandler) handleDevHome( c *gin.Context ) {
    udid := c.PostForm("udid")
    pc, udid := self.getPc( c )
    pc.doHome( udid )
}

func (self *DevHandler) handleDevSwipe( c *gin.Context ) {
    x1, _ := strconv.Atoi( c.PostForm("x1") )
    y1, _ := strconv.Atoi( c.PostForm("y1") )
    x2, _ := strconv.Atoi( c.PostForm("x2") )
    y2, _ := strconv.Atoi( c.PostForm("y2") )
    pc, udid := self.getPc( c )
    pc.doSwipe( udid, x1, y1, x2, y2 )
}

func (self *DevHandler) handleKeys( c *gin.Context ) {
    keys := c.PostForm("keys")
    pc, udid := self.getPc( c )
    pc.doKeys( udid, keys )
}

func (self *DevHandler) handleDevPing( c *gin.Context ) {
}

func (self *DevHandler) showDevVideo( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    
    dev := getDevice( udid )
    
    c.HTML( http.StatusOK, "devVideo", gin.H{
        "udid": udid,
        "clickWidth": dev.ClickWidth,
        "clickHeight": dev.ClickHeight,
    } )
}

func (self *DevHandler) handleDevStatus( c *gin.Context, ) {
    s := self.sessionManager.GetSession( c )
    
    session := self.sessionManager.session
    
    provider := session.Get( s, "provider" ).(ProviderOb)
        
    status := c.PostForm("status")
    
    fmt.Printf("devStatus request; status=%s\n", status )
    
    var ok struct {
        ok bool
    }
    ok.ok = true
    
    udid := c.PostForm("udid")
    fmt.Printf("  udid=%s\n", udid )
    if status == "exists" {
        fmt.Printf("Notified that device %s exists\n", udid )
        width, _       := strconv.Atoi( c.PostForm("width") )
        height, _      := strconv.Atoi( c.PostForm("height") )
        clickWidth, _  := strconv.Atoi( c.PostForm("clickWidth") )
        clickHeight, _ := strconv.Atoi( c.PostForm("clickHeight") )
        addDevice( udid, "unknown", provider.Id, width, height, clickWidth, clickHeight )
        self.devTracker.setDevProv( udid, provider.Id )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "info" {
        info := c.PostForm("info")
        fmt.Printf("Device info for %s:\n%s\n", udid, info )
        updateDeviceInfo( udid, info, provider.Id )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "wdaStarted" {
        fmt.Printf("WDA started for %s\n", udid )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "wdaStopped" {
        fmt.Printf("WDA stopped for %s\n", udid )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "provisionStopped" {
        fmt.Printf("Provision stopped for %s\n", udid )
        c.JSON( http.StatusOK, ok )
        return
    }
    
    var nok struct {
        ok bool
    }
    nok.ok = false
    c.JSON( http.StatusOK, nok )
}