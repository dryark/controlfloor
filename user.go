package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "github.com/gin-gonic/gin"
    cfauth "github.com/nanoscopic/controlfloor_auth"
    log "github.com/sirupsen/logrus"
)

type UserHandler struct {
    authHandler    cfauth.AuthHandler
    r              *gin.Engine
    devTracker     *DevTracker
    sessionManager *cfSessionManager
    config         *Config
}

func NewUserHandler(
    authHandler    cfauth.AuthHandler,
    r              *gin.Engine,
    devTracker     *DevTracker,
    sessionManager *cfSessionManager,
    config         *Config,
) *UserHandler {
    return &UserHandler{
        authHandler,
        r,
        devTracker,
        sessionManager,
        config,
    }
}

func (self *UserHandler) registerUserRoutes() (*gin.RouterGroup) {
    r := self.r
    
    fmt.Println("Registering user routes")
    r.GET("/login", self.showUserLogin )
    r.GET("/logout", self.handleUserLogout )
    r.POST("/login", self.handleUserLogin )
    uAuth := r.Group("/")
    uAuth.Use( self.NeedUserAuth( self.authHandler ) )
    uAuth.GET("/", self.showUserRoot )
    uAuth.GET("/imgStream", func( c *gin.Context ) {
        self.handleImgStream( c )
    } )
    return uAuth
}

func (self *UserHandler) NeedUserAuth( authHandler cfauth.AuthHandler ) gin.HandlerFunc {
    return func( c *gin.Context ) {
        sCtx := self.sessionManager.GetSession( c )
        
        loginI := self.sessionManager.session.Get( sCtx, "user" )
        
        if loginI == nil {
            if authHandler != nil {
                authHandler.UserAuth( c )
                return
            }
            
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

func (self *UserHandler) showUserRoot( c *gin.Context ) {
    devices, err := getDevices()
    if err != nil { panic( err ) }
    
    output := ""
    for _, device := range devices {
        output = output + fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td><a href="/devInfo?udid=%s">%s</a></td>
                <td>%d</td>
                <td>%s</td>
                <td>%d</td><td>%d</td><td>%d</td><td>%d</td>
            </tr>`,
            device.Name,
            device.Udid, device.Udid,
            device.ProviderId,
            device.JsonInfo,
            device.Width,
            device.Height,
            device.ClickWidth,
            device.ClickHeight,
        )
        // also Width, Height, ClickWidth, and ClickHeight
    }
    
    rs, _ := getReservations()
    if rs == nil {
        rs = make( map[string]DbReservation )
    }
    
    sCtx := self.sessionManager.GetSession( c )
    user := self.sessionManager.session.Get( sCtx, "user" ).(string)
    
    jsont := ""
    for _, device := range devices {
        udid := device.Udid
        
        provId := self.devTracker.getDevProvId( udid )
        if provId != 0 {
            r, hasR := rs[ udid ]
            if hasR && r.User != user {
                device.Ready = "In Use"
            } else {
                device.Ready = "Yes"
            }
        } else {
            device.Ready = "No"
        }
        
        t, _ := json.Marshal( device )
              
        jsont += string(t) + ","
    }
    if jsont != "" {
        jsont = jsont[:len( jsont )-1]
    }
    
    c.HTML( http.StatusOK, "userRoot", gin.H{
      "devices":      output,
      "devices_json": jsont,
      "deviceVideo":  self.config.text.deviceVideo,
    } )
}

func (self *UserHandler) showUserLogin( rCtx *gin.Context ) {
    rCtx.HTML( http.StatusOK, "userLogin", gin.H{} )
}

func (self *UserHandler) handleUserLogout( c *gin.Context ) {
    s := self.sessionManager.GetSession( c )
    
    self.sessionManager.session.Remove( s, "user" )
    self.sessionManager.WriteSession( c )
    
    c.Redirect( 302, "/" )
}

func (self *UserHandler) handleUserLogin( c *gin.Context ) {
    if self.authHandler != nil {
        success := self.authHandler.UserLogin( c )
        if success {
            c.Redirect( 302, "/" )
        } else {
            fmt.Printf("login failed\n")
            self.showUserLogin( c )
        }
        return
    }
    
    s := self.sessionManager.GetSession( c )
    
    user := c.PostForm("user")
    pass := c.PostForm("pass")
    
    if user == "ok" && pass == "ok" {
        fmt.Printf("login ok\n")
        
        self.sessionManager.session.Put( s, "user", "test" )
        self.sessionManager.WriteSession( c )
        
        c.Redirect( 302, "/" )
        return
    } else {
        fmt.Printf("login failed\n")
    }
    
    self.showUserLogin( c )
}

func (self *UserHandler) handleImgStream( c *gin.Context ) {
    //s := getSession( c )
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    rid, rok := c.GetQuery("rid")
    
    log.WithFields( log.Fields{
        "type": "imgstream_start",
        "udid": censorUuid( udid ),
        "rid": rid,
    } ).Info("Image stream connected")
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }
    
    stopChan := make( chan bool )
    
    self.devTracker.setVidStreamOutput( udid, &VidConn{
        socket: conn,
        stopChan: stopChan,
    } )
    
    fmt.Printf("sending startStream to provider\n")
    provId := self.devTracker.getDevProvId( udid )
    provConn := self.devTracker.getProvConn( provId )
    provConn.startImgStream( udid )
    
    <- stopChan
    
    log.WithFields( log.Fields{
        "type": "imgstream_start",
        "udid": censorUuid( udid ),
        "rid": rid,
    } ).Info("Image stream disconnected")
    
    if rok {
        deleteReservationWithRid( udid, rid )
    }
    provConn.stopImgStream( udid )
}