package main

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
)

func registerUserRoutes( r *gin.Engine, devTracker *DevTracker ) (*gin.RouterGroup) {
    fmt.Println("Registering user routes")
    r.GET("/login", showUserLogin )
    r.GET("/logout", handleUserLogout )
    r.POST("/login", handleUserLogin )
    uAuth := r.Group("/")
    uAuth.Use( NeedUserAuth() )
    uAuth.GET("/", showUserRoot )
    uAuth.GET("/imgStream", func( c *gin.Context ) {
        handleImgStream( c, devTracker )
    } )
    return uAuth
}

func NeedUserAuth() gin.HandlerFunc {
    return func( c *gin.Context ) {
        
        sCtx := getSession( c )
        
        loginI := session.Get( sCtx, "user" )
        
        if loginI == nil {
            
            c.Redirect( 302, "/login" )
            c.Abort()
            fmt.Println("user fail")
            return
        } else {
            //fmt.Println("user ok")
        }
        
        c.Next()
    }
}

func showUserRoot( c *gin.Context ) {
    var devices [] DbDevice
    err := gDb.Find( &devices )
    if err != nil {
        panic( err )
    }
    
    output := ""
    for _, device := range devices {
        output = output + fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td><a href="/devInfo?udid=%s">%s</a></td>
                <td>%d</td>
            </tr>`, device.Name, device.Udid, device.Udid, device.ProviderId )
    }
    
    c.HTML( http.StatusOK, "userRoot", gin.H{ "devices": output } )
}

func showUserLogin( rCtx *gin.Context ) {
    rCtx.HTML( http.StatusOK, "userLogin", gin.H{} )
}

func handleUserLogout( c *gin.Context ) {
    s := getSession( c )
    
    session.Remove( s, "user" )
    writeSession( c )
    
    c.Redirect( 302, "/" )
}

func handleUserLogin( c *gin.Context ) {
    s := getSession( c )
    
    /*loginI := session.Get( sCtx, "login" )
        
    if loginI == nil {
        showLogin
        return
    }*/
    user := c.PostForm("user")
    pass := c.PostForm("pass")
    
    if user == "ok" && pass == "ok" {
        fmt.Printf("login ok\n")
        
        //login := Login{
        //    level: 1,
        //}
        session.Put( s, "user", "test" )
        writeSession( c )
        
        c.Redirect( 302, "/" )
        //c.Data( http.StatusOK, "text/html", []byte("logged in") )
        return
    } else {
        fmt.Printf("login failed\n")
    }
    
    showUserLogin( c )
}

func handleImgStream( c *gin.Context, devTracker *DevTracker ) {
    //s := getSession( c )
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    
    fmt.Printf("connection to /imgStream udid=%s\n", udid )
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }
    
    stopChan := make( chan bool )
    
    devTracker.setVidStreamOutput( udid, &VidConn{
        socket: conn,
        stopChan: stopChan,
    } )
    
    fmt.Printf("sending startStream to provider\n")
    provId := devTracker.getDevProvId( udid )
    provConn := devTracker.getProvConn( provId )
    provConn.startImgStream( udid )
    
    <- stopChan
    
    provConn.stopImgStream( udid )
}