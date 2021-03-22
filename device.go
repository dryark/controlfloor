package main

import (
    "fmt"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
)

func registerDeviceRoutes( r *gin.Engine, pAuth *gin.RouterGroup, uAuth *gin.RouterGroup, devTracker *DevTracker ) {
    fmt.Println("Registering device routes")
    //pAuth.GET("/devStatus", showDevStatus )
    pAuth.POST("/devStatus", func( c *gin.Context ) {
        handleDevStatus( c, devTracker )
    } )
    // - Device is present on provider
    // - Device Info fetched from device
    // - WDA start/stop
    // - Video Streamer start/stop
    // - Video seems active/inactive
    
    //uAuth.GET("/devClick", showDevClick )
    uAuth.POST("/devClick",     func( c *gin.Context ) { handleDevClick( c, devTracker ) } )
    uAuth.POST("/devHardPress", func( c *gin.Context ) { handleDevHardPress( c, devTracker ) } )
    uAuth.POST("/devLongPress", func( c *gin.Context ) { handleDevLongPress( c, devTracker ) } )
    uAuth.POST("/devHome",      func( c *gin.Context ) { handleDevHome( c, devTracker ) } )
    uAuth.POST("/devSwipe",     func( c *gin.Context ) { handleDevSwipe( c, devTracker ) } )
    uAuth.POST("/keys",         func( c *gin.Context ) { handleKeys( c, devTracker ) } )
    
    uAuth.GET("/devInfo", func( c *gin.Context ) {
        showDevInfo( c, devTracker )
    } )
    
    uAuth.GET("/devVideo", showDevVideo )
    
    uAuth.GET("/devPing", handleDevPing )
}

func showDevInfo( c *gin.Context, devTracker *DevTracker ) {
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

func getPc( c *gin.Context, devTracker *DevTracker ) (*ProviderConnection,string) {
    udid := c.PostForm("udid")
    provId := devTracker.getDevProvId( udid )
    provConn := devTracker.getProvConn( provId )
    return provConn, udid
}

func handleDevClick( c *gin.Context, dt *DevTracker ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := getPc( c, dt )
    pc.doClick( udid, x, y )
}

func handleDevHardPress( c *gin.Context, dt *DevTracker ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := getPc( c, dt )
    pc.doHardPress( udid, x, y )
}

func handleDevLongPress( c *gin.Context, dt *DevTracker ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := getPc( c, dt )
    pc.doLongPress( udid, x, y )
}

func handleDevHome( c *gin.Context, dt *DevTracker ) {
    udid := c.PostForm("udid")
    pc, udid := getPc( c, dt )
    pc.doHome( udid )
}

func handleDevSwipe( c *gin.Context, dt *DevTracker ) {
    x1, _ := strconv.Atoi( c.PostForm("x1") )
    y1, _ := strconv.Atoi( c.PostForm("y1") )
    x2, _ := strconv.Atoi( c.PostForm("x2") )
    y2, _ := strconv.Atoi( c.PostForm("y2") )
    pc, udid := getPc( c, dt )
    pc.doSwipe( udid, x1, y1, x2, y2 )
}

func handleKeys( c *gin.Context, dt *DevTracker ) {
    keys := c.PostForm("keys")
    pc, udid := getPc( c, dt )
    pc.doKeys( udid, keys )
}

func handleDevPing( c *gin.Context ) {
}

func showDevVideo( c *gin.Context ) {
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

func handleDevStatus( c *gin.Context, devTracker *DevTracker ) {
    s := getSession( c )
    
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
        devTracker.setDevProv( udid, provider.Id )
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