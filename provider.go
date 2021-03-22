package main

import (
    "crypto/rand"
    "fmt"
    "encoding/hex"
    mrand "math/rand"
    "net/http"
    "sync"
    "time"
    //ecies "github.com/ecies/go"
    ws "github.com/gorilla/websocket"
    uj "github.com/nanoscopic/ujsonin/mod"
    "github.com/gin-gonic/gin"
)

func registerProviderRoutes( r *gin.Engine, devTracker *DevTracker ) (*gin.RouterGroup) {
    fmt.Println("Registering provider routes")
    r.POST("/register", handleRegister )
    r.GET("/provider/login", showProviderLogin )
    r.GET("/provider/logout", handleProviderLogout )
    r.POST("/provider/login", handleProviderLogin )
    
    pAuth := r.Group("/provider")
    pAuth.Use( NeedProviderAuth() )
    pAuth.GET("/", showProviderRoot )
    pAuth.GET("/ws", func( c *gin.Context ) {
        handleProviderWS( c, devTracker )
    } )
    pAuth.GET("/imgStream", func( c *gin.Context ) {
        handleImgProvider( c, devTracker )
    } )
    
    return pAuth
}

func NeedProviderAuth() gin.HandlerFunc {
    return func( c *gin.Context ) {
        sCtx := getSession( c )
        
        provider, ok := session.Get( sCtx, "provider" ).(ProviderOb)
        
        if !ok  {
            c.Redirect( 302, "/provider/login" )
            c.Abort()
            fmt.Println("provider fail")
            return
        } else {
            fmt.Printf("provider user=%s\n", provider.User )
        }
        
        c.Next()
    }
}

var wsupgrader = ws.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type ProvBase interface {  
    asText( int16 ) string
    needsResponse() bool
    resHandler() (func(*uj.JNode))
}

type ProvPing struct {
    blah string
    onRes func( *uj.JNode )
}
func (self *ProvPing) resHandler() (func(*uj.JNode) ) { return self.onRes }
func (self *ProvPing) needsResponse() (bool) { return true }
func (self *ProvPing) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"ping\"}\n", id)
}

type ProvClick struct {
    udid string
    x int
    y int
}
func (self *ProvClick) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvClick) needsResponse() (bool) { return false }
func (self *ProvClick) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"click\",udid:\"%s\",x:%d,y:%d}\n",id,self.udid,self.x,self.y)
}

type ProvHardPress struct {
    udid string
    x int
    y int
}
func (self *ProvHardPress) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvHardPress) needsResponse() (bool) { return false }
func (self *ProvHardPress) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"hardPress\",udid:\"%s\",x:%d,y:%d}\n",id,self.udid,self.x,self.y)
}

type ProvLongPress struct {
    udid string
    x int
    y int
}
func (self *ProvLongPress) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvLongPress) needsResponse() (bool) { return false }
func (self *ProvLongPress) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"longPress\",udid:\"%s\",x:%d,y:%d}\n",id,self.udid,self.x,self.y)
}

type ProvHome struct {
    udid string
}
func (self *ProvHome) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvHome) needsResponse() (bool) { return false }
func (self *ProvHome) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"home\",udid:\"%s\"}\n",id,self.udid)
}

type ProvKeys struct {
    udid string
    keys string
}
func (self *ProvKeys) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvKeys) needsResponse() (bool) { return false }
func (self *ProvKeys) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"keys\",udid:\"%s\",keys:\"%s\"}\n",id,self.udid,self.keys)
}

type ProvSwipe struct {
    udid string
    x1 int
    y1 int
    x2 int
    y2 int
}
func (self *ProvSwipe) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvSwipe) needsResponse() (bool) { return false }
func (self *ProvSwipe) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"swipe\",udid:\"%s\",x1:%d,y1:%d,x2:%d,y2:%d}\n",id,self.udid,self.x1,self.y1,self.x2,self.y2)
}

type ProvStartStream struct {
    udid string
}
func (self *ProvStartStream) resHandler() (func(*uj.JNode) ) { return nil }
func (self *ProvStartStream) needsResponse() (bool) { return false }
func (self *ProvStartStream) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"startStream\",udid:\"%s\"}\n",id,self.udid)
}

type ProvStopStream struct {
    udid string
}

func (self *ProvStopStream) resHandler() (func(*uj.JNode) ) {
    return nil
}

func (self *ProvStopStream) asText( id int16 ) (string) {
    return fmt.Sprintf("{id:%d,type:\"stopStream\",udid:\"%s\"}\n",id,self.udid)
}

func (self *ProvStopStream) needsResponse() (bool) {
    return false
}

type ReqTracker struct {
    reqMap map[int16] ProvBase
    lock *sync.Mutex
    conn *ws.Conn
}

func NewReqTracker() (*ReqTracker) {
    self := &ReqTracker{
        reqMap: make( map[int16] ProvBase ),
        lock: &sync.Mutex{},
    }
    
    return self
}

func (self *ReqTracker) sendReq( req ProvBase ) (error) {
    var reqText string
    if req.needsResponse() {
        var id int16
        maxi := ^uint16(0) / 2
        for {
            id = int16( mrand.Int31n( int32(maxi-2) ) ) + 1
            _, exists := self.reqMap[ id ]
            if !exists { break }
        }
        
        self.lock.Lock()
        self.reqMap[ id ] = req
        self.lock.Unlock()
        reqText = req.asText( id )
    } else {
        reqText = req.asText( 0 )
    }
    
    fmt.Printf("sending %s\n", reqText )
    // send the request
    err := self.conn.WriteMessage( ws.TextMessage, []byte(reqText) )
    return err
}
func (self *ReqTracker) processResp( msgType int, reqText []byte ) {
    fmt.Printf( "received %s\n", string(reqText) )
    
    if len( reqText ) == 0 {
        return
    }
    c1 := string( []byte{ reqText[0] } )
    if c1 != "{" {
        return
    }
    last := string( []byte{ reqText[ len( reqText ) - 2 ] } )
    if last != "}" {
        fmt.Printf("respond not json\n")
        return
    }
    
    root, _ := uj.Parse( reqText )
    id := root.Get("id").Int()
    
    req := self.reqMap[ int16(id) ]
    resHandler := req.resHandler()
    if resHandler != nil {
        resHandler( root )
    }
    
    self.lock.Lock()
    delete( self.reqMap, int16(id) )
    self.lock.Unlock()
    // deserialize the reqText to get the id
    // fetch the original request from the reqMap
    // respond to the original request if needed
}

type ProviderConnection struct {
    provChan chan ProvBase
    reqTracker *ReqTracker
}

func NewProviderConnection( provChan chan ProvBase ) (*ProviderConnection) {
    self := &ProviderConnection{
        provChan: provChan,
        reqTracker: NewReqTracker(),
    }
    
    return self
}

func (self *ProviderConnection) doPing() {
    ping := &ProvPing{
        onRes: func( root *uj.JNode ) {
            text := root.Get("text").String()
            fmt.Printf("pong text %s\n", text )
        },
    }
    self.provChan <- ping
}

func (self *ProviderConnection) doClick( udid string, x int, y int ) {
    click := &ProvClick{
        udid: udid,
        x: x,
        y: y,
    }
    self.provChan <- click
}

func (self *ProviderConnection) doHardPress( udid string, x int, y int ) {
    click := &ProvHardPress{
        udid: udid,
        x: x,
        y: y,
    }
    self.provChan <- click
}

func (self *ProviderConnection) doLongPress( udid string, x int, y int ) {
    click := &ProvLongPress{
        udid: udid,
        x: x,
        y: y,
    }
    self.provChan <- click
}

func (self *ProviderConnection) doHome( udid string ) {
    home := &ProvHome{
        udid: udid,
    }
    self.provChan <- home
}

func (self *ProviderConnection) doKeys( udid string, keys string ) {
    action := &ProvKeys{
        udid: udid,
        keys: keys,
    }
    self.provChan <- action
}

func (self *ProviderConnection) doSwipe( udid string, x1 int, y1 int, x2 int, y2 int ) {
    swipe := &ProvSwipe{
        udid: udid,
        x1: x1,
        y1: y1,
        x2: x2,
        y2: y2,
    }
    self.provChan <- swipe
}

func (self *ProviderConnection) startImgStream( udid string ) {
    self.provChan <- &ProvStartStream{ udid: udid }
}

func (self *ProviderConnection) stopImgStream( udid string ) {
    self.provChan <- &ProvStopStream{ udid: udid }
}

func handleImgProvider( c *gin.Context, devTracker *DevTracker ) {
    //s := getSession( c )
    
    //provider := session.Get( s, "provider" ).(ProviderOb)
    
    udid, uok := c.GetQuery("udid")
    if !uok {
        c.HTML( http.StatusOK, "error", gin.H{
            "text": "no uuid set",
        } )
        return
    }
    fmt.Printf("connection to provider/imgStream udid=%s\n", udid )
    
    //dev := getDevice( udid )
    
    provId := devTracker.getDevProvId( udid )
    provConn := devTracker.getProvConn( provId )
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }
    
    vidConn := devTracker.getVidStreamOutput( udid )
    outSocket := vidConn.socket
    
    go func() {
        for {
            if _, _, err := outSocket.NextReader(); err != nil {
                outSocket.Close()
                break
            }
        }
    }()
    for {
        t, data, err := conn.ReadMessage()
        if err != nil {
            conn = nil
            break
        }
        err = outSocket.WriteMessage( t, data )
        if err != nil {
            outSocket = nil
            provConn.stopImgStream( udid )
            break
        }
    }
    
    if conn != nil { conn.Close() }
    if outSocket != nil { outSocket.Close() }
}

func handleProviderWS( c *gin.Context, devTracker *DevTracker ) {
    s := getSession( c )
    
    provider := session.Get( s, "provider" ).(ProviderOb)
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }

    provChan := make( chan ProvBase )
    provConn := NewProviderConnection( provChan )
    devTracker.setProvConn( provider.Id, provConn )
    reqTracker := provConn.reqTracker
    reqTracker.conn = conn
    
    amDone := false
    
    fmt.Printf("got ws connection\n")
    
    go func() { for {
        time.Sleep( time.Second * 5 )
        provConn.doPing()
        //fmt.Printf("triggered periodic ping\n")
        if amDone { break }
    } }()
    
    go func() { for {
        t, msg, err := conn.ReadMessage()
        if err != nil {
            amDone = true
        }
        reqTracker.processResp( t, msg )
        
        if amDone { break }
    } }()
        
    for {
        ev := <- provChan
        err := reqTracker.sendReq( ev )
        if err != nil {
            break
        }        
    }
    
    devTracker.clearProvConn( provider.Id )
    fmt.Printf("lost ws connection\n")
}

func randHex() (string) {
    c := 16
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString( b )
}

func handleRegister( c *gin.Context ) {
    pass := c.PostForm("regPass")
    
    conf := getConf()
    if pass != conf.RegPass {
        var jsonf struct {
            Success bool
        }
        jsonf.Success = false
        c.JSON( http.StatusOK, jsonf )
        return
    }
    
    username := c.PostForm("username")
    
    var json struct {
        Success bool
        Password string
        Existed bool
    }
    json.Success = true
    pPass := randHex()
    json.Password = pPass
    existed := addProvider( username, pPass ) 
    json.Existed = existed
    
    c.JSON( http.StatusOK, json )
}

type ProviderOb struct {
    User string
    Id int64
}

func handleProviderLogin( c *gin.Context ) {
    s := getSession( c )
    
    user := c.PostForm("user")
    pass := c.PostForm("pass")
    fmt.Printf("Provider login user=%s pass=%s\n", user, pass )
    
    // ensure the user is legit
    provider := getProvider( user )
    if provider == nil {
        fmt.Printf("provider login failed 1\n")
        c.Redirect( 302, "/provider/?fail=1" )
        return
    }
    
    if pass == provider.Password {
        fmt.Printf("provider login ok\n")
        
        session.Put( s, "provider", &ProviderOb{
            User: user,
            Id: provider.Id,
        } )
        writeSession( c )
        
        c.Redirect( 302, "/provider/" )
        return
    } else {
        fmt.Printf("provider login failed [submit]%s != [db]%s\n", pass, provider.Password)
        c.Redirect( 302, "/provider/?fail=2" )
        return
    }
    
    showProviderLogin( c )
}

func handleProviderLogout( c *gin.Context ) {
    s := getSession( c )
    
    session.Remove( s, "provider" )
    writeSession( c )
    
    c.Redirect( 302, "/" )
}

func showProviderLogin( rCtx *gin.Context ) {
    rCtx.HTML( http.StatusOK, "providerLogin", gin.H{} )
}

func showProviderRoot( c *gin.Context ) {
    c.HTML( http.StatusOK, "providerRoot", gin.H{} )
} 