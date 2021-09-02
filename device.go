package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    "math/rand"
    "github.com/gin-gonic/gin"
    uj "github.com/nanoscopic/ujsonin/v2/mod"
    log "github.com/sirupsen/logrus"
    ws "github.com/gorilla/websocket"
)

type DevHandler struct {
    providerAuthGroup *gin.RouterGroup
    userAuthGroup     *gin.RouterGroup
    devTracker        *DevTracker
    sessionManager    *cfSessionManager
    config            *Config
}

func NewDevHandler(
    providerAuthGroup *gin.RouterGroup,
    userAuthGroup     *gin.RouterGroup,
    devTracker        *DevTracker,
    sessionManager    *cfSessionManager,
    config            *Config,
) *DevHandler {
    return &DevHandler{
        providerAuthGroup,
        userAuthGroup,
        devTracker,
        sessionManager,
        config,
    }
}

func (self *DevHandler) registerDeviceRoutes() {
    pAuth := self.providerAuthGroup
    uAuth := self.userAuthGroup
    
    fmt.Println("Registering device routes")
    pAuth.POST("/device/status/:variant", func( c *gin.Context ) {
        self.handleDevStatus( c )
    } )
    // - Device is present on provider
    // - Device Info fetched from device
    // - WDA start/stop
    // - Video Streamer start/stop
    // - Video seems active/inactive
    
    //uAuth.GET("/devClick", showDevClick )
    uAuth.POST("/device/click",     func( c *gin.Context ) { self.handleDevClick( c ) } )
    uAuth.POST("/device/hardPress", func( c *gin.Context ) { self.handleDevHardPress( c ) } )
    uAuth.POST("/device/longPress", func( c *gin.Context ) { self.handleDevLongPress( c ) } )
    uAuth.POST("/device/home",      func( c *gin.Context ) { self.handleDevHome( c ) } )
    uAuth.POST("/device/swipe",     func( c *gin.Context ) { self.handleDevSwipe( c ) } )
    uAuth.POST("/device/keys",      func( c *gin.Context ) { self.handleKeys( c ) } )
    uAuth.POST("/device/source",    func( c *gin.Context ) { self.handleSource( c ) } )
    uAuth.POST("/device/shutdown",  func( c *gin.Context ) { self.handleShutdown( c ) } )
      
    uAuth.GET("/device/info",       func( c *gin.Context ) { self.showDevInfo( c ) } )
    uAuth.GET("/device/info/json",  func( c *gin.Context ) { self.showDevInfoJson( c ) } )
    
    uAuth.GET("/device/imgStream",  func( c *gin.Context ) { self.handleImgStream( c ) } )
    uAuth.GET("/device/ws",         func( c *gin.Context ) { self.handleDevWs( c ) } )
    
    uAuth.GET("/device/video", self.showDevVideo )
    uAuth.GET("/device/reserved", self.showDevReservedTest )
    uAuth.GET("/device/kick", self.devKick )
    uAuth.POST("/device/videoStop", self.stopDevVideo )
    
    uAuth.GET("/device/ping", self.handleDevPing )
    uAuth.GET("/device/inspect", self.showDevInspect )
}

type SRawInfo struct {
    ArtworkDeviceProductDescription string `json:"ArtworkDeviceProductDescription" example:"iPhone 12"`
    DeviceName string `json:"DeviceName" example:"iPhone"`
    EthernetAddress string `json:"EthernetAddress" example:"b0:8c:75:75:aa:a4"`
    HardwareModel string `json:"HardwareModel" example:"D53gAP"`
    InternationalMobileEquipmentIdentity string `json:"InternationalMobileEquipmentIdentity" example:"355727333663572"`
    ModelNumber string `json:"ModelNumber" example:"MGH63"`
    ProductType string `json:"ProductType" example:"iPhone13,2"`
    ProductVersion string `json:"ProductVersion" example:"14.2.1"`
    UniqueDeviceID string `json:"UniqueDeviceID" example:"00008100-001338811EE10033"`
}

type SDeviceInfoFail struct {
    Success     bool `json:"success" example:"false"`
    Err         string `json:"error" example:"some error"`
}

type SDeviceInfo struct {
    Udid        string `json:"udid"        example:"00008100-001338811EE10033"`
    Name        string `json:"name"        example:"Phone Name"`
    ClickWidth  int    `json:"clickWidth"  example:"390"`
    ClickHeight int    `json:"clickHeight" example:"844"`
    VidWidth    int    `json:"vidWidth"    example:"390"`
    VidHeight   int    `json:"vidHeight"   example:"844"`
    Provider    int    `json:"provider"    example:"1"`
    RawInfo     string `json:"rawInfo"`
    WdaStatus   string `json:"wdaStatus"   example:"up"`
    VideoStatus string `json:"videoStatus" example:"up"`
    DeviceVideo string `json:"deviceVideo" example:"up"`
}

// @Summary Device - Device info JSON
// @Router /device/info/json [GET]
// @Param udid query string true "Device UDID"
// @Produce json
// @Success 200 {object} SDeviceInfo
func (self *DevHandler) showDevInfoJson( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.JSON( http.StatusOK, SDeviceInfoFail{
            Success: false,
            Err: "Must pass udid",
        } )
        return
    }
    
    dev := getDevice( udid )
    if dev == nil {
        c.JSON( http.StatusOK, SDeviceInfoFail{
            Success: false,
            Err: "No device with that udid",
        } )
        return
    }
    
    info := dev.JsonInfo
    
    stat := self.devTracker.getDevStatus( udid )
    wdaUp := "-"
    videoUp := "-"
    if stat != nil {
        wdaUp = "up"
        if !stat.wda { wdaUp = "down" }
        videoUp = "up"
        if !stat.video { videoUp = "down" }
    }
    
    provId := self.devTracker.getDevProvId( udid )
    
    c.JSON( http.StatusOK, SDeviceInfo{
        Udid:        udid,
        Name:        dev.Name,
        ClickWidth:  dev.ClickWidth,
        ClickHeight: dev.ClickHeight,
        VidWidth:    dev.Width,
        VidHeight:   dev.Height,
        Provider:    int(provId),
        RawInfo:     info,
        WdaStatus:   wdaUp,
        VideoStatus: videoUp,
        DeviceVideo: self.config.text.deviceVideo,
    } )
}

// @Summary Device - Device info page
// @Router /device/info [GET]
// @Param udid query string true "Device UDID"
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
    
    info := dev.JsonInfo
    if info != "" {
      var obj map[string]interface{}
      json.Unmarshal([]byte(info), &obj)
      infoBytes, _ := json.MarshalIndent(obj, "<br>", " &nbsp; &nbsp; &nbsp; ")
      info = string( infoBytes )
    }
    
    stat := self.devTracker.getDevStatus( udid )
    wdaUp := "-"
    videoUp := "-"
    if stat != nil {
        wdaUp = "up"
        if !stat.wda { wdaUp = "down" }
        videoUp = "up"
        if !stat.video { videoUp = "down" }
    }
    
    provId := self.devTracker.getDevProvId( udid )
    
    c.HTML( http.StatusOK, "devInfo", gin.H{
        "udid":        udid,
        "name":        dev.Name,
        "clickWidth":  dev.ClickWidth,
        "clickHeight": dev.ClickHeight,
        "vidWidth":    dev.Width,
        "vidHeight":   dev.Height,
        "provider":    provId,
        "info":        info,
        "wdaStatus":   wdaUp,
        "videoStatus": videoUp,
        "deviceVideo": self.config.text.deviceVideo,
    } )
}

func (self *DevHandler) getPc( c *gin.Context ) (*ProviderConnection,string) {
    udid := c.PostForm("udid")
    provId := self.devTracker.getDevProvId( udid )
    provConn := self.devTracker.getProvConn( provId )
    return provConn, udid
}

// @Summary Device - Click coordinate
// @Router /device/click [POST]
// @Param udid formData string true "Device UDID"
// @Param x formData int true "x"
// @Param y formData int true "y"
func (self *DevHandler) handleDevClick( c *gin.Context ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := self.getPc( c )
    
    done := make( chan bool )
    
    pc.doClick( udid, x, y, func( uj.JNode, []byte ) {
        done <- true
    } )
    
    <- done
    
    c.HTML( http.StatusOK, "error", gin.H{
        "text": "ok",
    } )
}

// @Summary Device - Hard press coordinate
// @Router /device/hardPress [POST]
// @Param udid formData string true "Device UDID"
// @Param x formData int true "x"
// @Param y formData int true "y"
func (self *DevHandler) handleDevHardPress( c *gin.Context ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := self.getPc( c )
    pc.doHardPress( udid, x, y )
}

// @Summary Device - Long Press coordinate
// @Router /device/longPress [POST]
// @Param udid formData string true "Device UDID"
// @Param x formData int true "x"
// @Param y formData int true "y"
func (self *DevHandler) handleDevLongPress( c *gin.Context ) {
    x, _ := strconv.Atoi( c.PostForm("x") )
    y, _ := strconv.Atoi( c.PostForm("y") )
    pc, udid := self.getPc( c )
    pc.doLongPress( udid, x, y )
}

// @Summary Device click
// @Router /device/home [POST]
// @Param udid formData string true "Device UDID"
func (self *DevHandler) handleDevHome( c *gin.Context ) {
    //udid := c.PostForm("udid")
    pc, udid := self.getPc( c )
    
    done := make( chan bool )
    
    pc.doHome( udid, func( uj.JNode, []byte ) {
        done <- true
    } )
    
    <- done
    
    c.HTML( http.StatusOK, "error", gin.H{
        "text": "ok",
    } )
}

// @Summary Device - Swipe
// @Router /device/swipe [POST]
// @Param udid formData string true "Device UDID"
// @Param x1 formData int true "x1"
// @Param y1 formData int true "y1"
// @Param x2 formData int true "x2"
// @Param y2 formData int true "y2"
// @Param delay formData number true "Time of swipe"
func (self *DevHandler) handleDevSwipe( c *gin.Context ) {
    x1, _ := strconv.Atoi( c.PostForm("x1") )
    y1, _ := strconv.Atoi( c.PostForm("y1") )
    x2, _ := strconv.Atoi( c.PostForm("x2") )
    y2, _ := strconv.Atoi( c.PostForm("y2") )
    delay, _ := strconv.ParseFloat( c.PostForm("delay"), 64 )
    pc, udid := self.getPc( c )
    
    done := make( chan bool )
    
    pc.doSwipe( udid, x1, y1, x2, y2, delay, func( uj.JNode, []byte ) {
        done <- true
    } )
    
    <- done
    
    c.HTML( http.StatusOK, "error", gin.H{
        "text": "ok",
    } )
}

// @Summary Device - Simulate keystrokes
// @Router /device/keys [POST]
// @Param udid formData string true "Device UDID"
// @Param curid formData int true "Incrementing unique ID"
// @Param keys formData string true "Keys"
// @Param prevkeys formData string true "Previous keys"
func (self *DevHandler) handleKeys( c *gin.Context ) {
    keys     := c.PostForm("keys")
    curid, _ := strconv.Atoi( c.PostForm("curid") )
    prevkeys := c.PostForm("prevkeys")
    
    done := make( chan bool )
    
    pc, udid := self.getPc( c )
    pc.doKeys( udid, keys, curid, prevkeys, func( uj.JNode, []byte ) {
        done <- true
    } )
    
    <- done
    
    c.HTML( http.StatusOK, "error", gin.H{
        "text": "ok",
    } )
}

// @Summary Device - Get device source
// @Router /device/source [GET]
// @Param udid formData string true "Device UDID"
func (self *DevHandler) handleSource( c *gin.Context ) {
    pc, udid := self.getPc( c )
    
    done := make( chan bool )
    
    pc.doSource( udid, func( _ uj.JNode, raw []byte ) {
        c.Writer.Header().Set("Content-Type", "text/json; charset=utf-8")
        c.Writer.WriteHeader(200)
        c.Writer.Write( raw )
        done <- true
    } )
    
    <- done
}

// @Summary Device - Shutdown device provider
// @Router /device/shutdown [GET]
// @Param udid formData string true "Device UDID"
func (self *DevHandler) handleShutdown( c *gin.Context ) {
    pc, udid := self.getPc( c )
        
    pc.doShutdown( func( _ uj.JNode, raw []byte ) {} )
    self.devTracker.clearDevProv( udid )
    
    // It will take at least 3 seconds to restart
    time.Sleep( time.Second * 3 )
    
    // wait for the device with the specified UDID to return
    i := 0
    for {
        i++
        provId := self.devTracker.getDevProvId( udid )
        if provId == 0 {
            if i == 30 { break }
            time.Sleep( time.Second )
            continue
        }
        provConn := self.devTracker.getProvConn( provId )
        if provConn == nil {
            if i == 30 { break }
            time.Sleep( time.Second )
            continue
        }
        status := self.devTracker.getDevStatus( udid )
        if status.video == false {
            if i == 30 { break }
            time.Sleep( time.Second )
            continue
        }
        c.Writer.Header().Set("Content-Type", "text/json; charset=utf-8")
        c.Writer.WriteHeader(200)
        c.Writer.Write( []byte("{success:true}") )
        return
    }
    
    c.Writer.Header().Set("Content-Type", "text/json; charset=utf-8")
    c.Writer.WriteHeader(200)
    c.Writer.Write( []byte("{success:false}") )
}

func (self *DevHandler) handleDevPing( c *gin.Context ) {
}

// @Summary Device - Kick device user
// @Router /device/kick [GET]
// @Param udid query string true "Device UDID"
func (self *DevHandler) devKick( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    
    self.devTracker.msgClient( udid, ClientMsg{ msgType: CMKick, msg: "{\"type\":\"kick\"}" } )
    
    deleteReservation( udid )
    
    c.Redirect( 302, "/devVideo?udid=" + udid )
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func (self *DevHandler) showDevReservedTest( c *gin.Context ) {
    udid, _ := c.GetQuery("udid")
    c.HTML( http.StatusOK, "devReserved", gin.H{
        "udid": udid,
        "user": "some user",
    } )
}

// @Summary Device - Video Page
// @Router /device/video [GET]
// @Param udid query string true "Device UDID"
func (self *DevHandler) showDevVideo( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    
    dev := getDevice( udid )
    
    sCtx := self.sessionManager.GetSession( c )
    user := self.sessionManager.session.Get( sCtx, "user" ).(string)
    fmt.Printf("Reserving device %s for %s\n", udid, user )
    rid := RandStringBytes( 10 )
    success := addReservation( udid, user, rid )
    
    if !success {
        rv := getReservation( udid )
        
        if rv.User != user {
            c.HTML( http.StatusOK, "devReserved", gin.H{
                "udid": udid,
                "user": rv.User,
            } )
            return
        }
        fmt.Printf("Renewing reservation\n")
        deleteReservation( udid )
        addReservation( udid, user, rid )
    }
    
    rawInfo := dev.JsonInfo
    info := "{}"
    if rawInfo != "" {
      var obj map[string]interface{}
        json.Unmarshal([]byte(rawInfo), &obj)
        infoBytes, _ := json.MarshalIndent(obj, "<br>", " &nbsp; &nbsp; &nbsp; ")
        info = string( infoBytes )
    }
    
    notesText := "{}"
    if self.config.notes != nil {
        notesText = self.config.notes.JsonSave()
    }
    
    c.HTML( http.StatusOK, "devVideo", gin.H{
        "udid":        udid,
        "clickWidth":  dev.ClickWidth,
        "clickHeight": dev.ClickHeight,
        "vidWidth":    dev.Width,
        "vidHeight":   dev.Height,
        "rid":         rid,
        "idleTimeout": self.devTracker.config.idleTimeout,
        "maxHeight":   self.config.maxHeight,
        "deviceVideo": self.config.text.deviceVideo,
        "info":        info,
        "rawInfo":     rawInfo,
        "notes":       notesText,
    } )
}

// @Summary Device - Inspect Page
// @Router /device/inspect [GET]
// @Param udid query string true "Device UDID"
func (self *DevHandler) showDevInspect( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    
    dev := getDevice( udid )
    
    c.HTML( http.StatusOK, "devInspect", gin.H{
        "udid": udid,
        "vidWidth": dev.Width,
        "vidHeight": dev.Height,
    } )
}

// @Summary Device - Stop device video
// @Router /device/videoStop [POST]
// @Param udid query string true "Device UDID"
func (self *DevHandler) stopDevVideo( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    rid, rok := c.GetQuery("rid")
    if !rok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no rid set",
        } )
        return
    }
    
    fmt.Printf("dev video stopped for udid: %s\n", udid )
    
    deleteReservationWithRid( udid, rid )
    
    c.HTML( http.StatusOK, "error", gin.H{
        "text": "ok",
    } )
}

// @Summary Device Status - Existence
// @Router /device/status/exists [POST]
// @Param udid query string true "Device UDID"
func dummy1() {}

// @Summary Device Status - Information
// @Router /device/status/info [POST]
// @Param udid query string true "Device UDID"
func dummy2() {}

// @Summary Device Status - WDA Started
// @Router /device/status/wdaStarted [POST]
// @Param udid query string true "Device UDID"
func dummy4() {}

// @Summary Device Status - WDA Stopped
// @Router /device/status/wdaStopped [POST]
// @Param udid query string true "Device UDID"
func dummy5() {}

// @Summary Device Status - Video Started
// @Router /device/status/videoStarted [POST]
// @Param udid query string true "Device UDID"
func dummy6() {}

// @Summary Device Status - Video Stopped
// @Router /device/status/videoStopped [POST]
// @Param udid query string true "Device UDID"
func dummy7() {}

// @Summary Device Status - Provision Stopped
// @Router /provider/device/status/provisionStopped [POST]
// @Param udid query string true "Device UDID"

func (self *DevHandler) handleDevStatus( c *gin.Context, ) {
    s := self.sessionManager.GetSession( c )
    
    session := self.sessionManager.session
    
    provider := session.Get( s, "provider" ).(ProviderOb)
        
    //status := c.PostForm("status")
    variant := c.Param("variant")
    
    fmt.Printf("devStatus request; variant=%s\n", variant )
    
    var ok struct {
        ok bool
    }
    ok.ok = true
    
    udid := c.PostForm("udid")
    fmt.Printf("  udid=%s\n", udid )

    if variant == "exists" {
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
    if variant == "info" {
        info := c.PostForm("info")
        fmt.Printf("Device info for %s:\n%s\n", udid, info )
        updateDeviceInfo( udid, info, provider.Id )
        c.JSON( http.StatusOK, ok )
        return
    }
    if variant == "wdaStarted" {
        fmt.Printf("WDA started for %s\n", udid )
        self.devTracker.setDevStatus( udid, "wda", true )
        c.JSON( http.StatusOK, ok )
        return
    }
    if variant == "wdaStopped" {
        fmt.Printf("WDA stopped for %s\n", udid )
        self.devTracker.setDevStatus( udid, "wda", false )
        c.JSON( http.StatusOK, ok )
        return
    }
    if variant == "videoStarted" {
        fmt.Printf("Video started for %s\n", udid )
        self.devTracker.setDevStatus( udid, "video", true )
        c.JSON( http.StatusOK, ok )
        return
    }
    if variant == "videoStopped" {
        fmt.Printf("Video stopped for %s\n", udid )
        self.devTracker.setDevStatus( udid, "video", false )
        c.JSON( http.StatusOK, ok )
        return
    }
    if variant == "provisionStopped" {
        fmt.Printf("Provision stopped for %s\n", udid )
        self.devTracker.clearDevProv( udid )
        c.JSON( http.StatusOK, ok )
        return
    }
    
    var nok struct {
        ok bool
    }
    nok.ok = false
    c.JSON( http.StatusOK, nok )
}

// @Description Device - Image Stream Websocket
// @Router /device/imgStream [GET]
// @Param udid query string true "Device UDID"
// @Param rid query string true "Video Instance ID"
func (self *DevHandler) handleImgStream( c *gin.Context ) {
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
    if provId == 0 {
        fmt.Println("Device not yet provided")
        return
    }
    provConn := self.devTracker.getProvConn( provId )
    if provConn == nil {
        fmt.Println("Device not yet provided")
        return
    }
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

type WsResponse interface {
    String() string
}

type SyncResponse struct {
    id int
}

func ( self SyncResponse ) String() string {
    return ""
}

// @Description Device - Device Command Websocket
// @Router /device/ws [GET]
// @Param udid query string true "Device UDID"
func (self *DevHandler) handleDevWs( c *gin.Context ) {
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    
    log.WithFields( log.Fields{
        "type": "devws_start",
        "udid": censorUuid( udid ),
    } ).Info("Device ws connected")
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }
    
    for {
        t, msg, err := conn.ReadMessage()
        if err != nil {
            fmt.Printf("Error reading from ws\n")
            break
        }
        if t == ws.TextMessage {
            //tMsg := string( msg )
            b1 := []byte{ msg[0] }
            if string(b1) == "{" {
                root, _ := uj.Parse( msg )
                id := root.Get("id").Int()
                mType := root.Get("type").String()
                var resp WsResponse
                if mType == "timesync" {
                    resp = SyncResponse{id:id}
                }
                if resp != nil {
                    respStr := resp.String()
                    err := conn.WriteMessage( ws.TextMessage, []byte( respStr ) )
                    if err != nil {
                        fmt.Printf("Error writing to ws\n")
                    }
                }
            }
        }
    }
}