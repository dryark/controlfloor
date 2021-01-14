package main

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
)

func registerDeviceRoutes( r *gin.Engine, pAuth *gin.RouterGroup, uAuth *gin.RouterGroup ) {
    //pAuth.GET("/devStatus", showDevStatus )
    pAuth.POST("/devStatus", handleDevStatus )
    // - Device is present on provider
    // - Device Info fetched from device
    // - WDA start/stop
    // - Video Streamer start/stop
    // - Video seems active/inactive
    
    //uAuth.GET("/devClick", showDevClick )
    //uAuth.POST("/devClick", handleDevClick )
    
    //uAuth.GET("/devInfo", showDevInfo )
    
    uAuth.GET("/devVideo", showDevVideo )
}

func showDevVideo( c *gin.Context ) {
    var devices [] DbDevice
    err := gDb.Find( &devices )
    if err != nil {
        panic( err )
    }
    
    output := ""
    for _, device := range devices {
        output = output + fmt.Sprintf("Name: %s<br>\n", device.Name )
    }
    
    c.HTML( http.StatusOK, "devVideo", gin.H{} )
}

func handleDevStatus( c *gin.Context ) {
    status := c.PostForm("status")
    
    var ok struct {
        ok bool
    }
    ok.ok = true
    
    if status == "exists" {
        uuid := c.PostForm("uuid")
        fmt.Printf("Notified that device %s exists\n", uuid )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "info" {
        uuid := c.PostForm("uuid")
        info := c.PostForm("info")
        fmt.Printf("Device info for %s:\n%s\n", uuid, info )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "wdaStarted" {
        uuid := c.PostForm("uuid")
        fmt.Printf("WDA started for %s\n", uuid )
        c.JSON( http.StatusOK, ok )
        return
    }
    if status == "provisionStopped" {
        uuid := c.PostForm("uuid")
        fmt.Printf("Provision stopped for %s\n", uuid )
        c.JSON( http.StatusOK, ok )
        return
    }
    
    var nok struct {
        ok bool
    }
    nok.ok = false
    c.JSON( http.StatusOK, nok )
}