package main

import (
    "crypto/rand"
    "fmt"
    "encoding/hex"
    mrand "math/rand"
    "net/http"
    "sync"
    "time"
    ws "github.com/gorilla/websocket"
    uj "github.com/nanoscopic/ujsonin/mod"
    "github.com/gin-gonic/gin"
)

type ProviderHandler struct {
    r              *gin.Engine
    devTracker     *DevTracker
    sessionManager *cfSessionManager
}

func NewProviderHandler(
    r              *gin.Engine,
    devTracker     *DevTracker,
    sessionManager *cfSessionManager,
) *ProviderHandler {
    return &ProviderHandler{
        r,
        devTracker,
        sessionManager,
    }
}

func (self *ProviderHandler) registerProviderRoutes() (*gin.RouterGroup) {
    r := self.r
    
    fmt.Println("Registering provider routes")
    r.POST("/register", self.handleRegister )
    r.GET("/provider/login", self.showProviderLogin )
    r.GET("/provider/logout", self.handleProviderLogout )
    r.POST("/provider/login", self.handleProviderLogin )
    
    pAuth := r.Group("/provider")
    pAuth.Use( self.NeedProviderAuth() )
    pAuth.GET("/", self.showProviderRoot )
    pAuth.GET("/ws", func( c *gin.Context ) {
        self.handleProviderWS( c )
    } )
    pAuth.GET("/imgStream", func( c *gin.Context ) {
        self.handleImgProvider( c )
    } )
    
    return pAuth
}

func (self *ProviderHandler) NeedProviderAuth() gin.HandlerFunc {
    return func( c *gin.Context ) {
        sCtx := self.sessionManager.GetSession( c )
        
        provider, ok := self.sessionManager.session.Get( sCtx, "provider" ).(ProviderOb)
        
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

const (
    CMKick = iota
)
type ClientMsg struct {
    msgType int
    msg     string
}

func (self *ProviderHandler) handleImgProvider( c *gin.Context ) {
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
    
    provId := self.devTracker.getDevProvId( udid )
    provConn := self.devTracker.getProvConn( provId )
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }
    
    vidConn := self.devTracker.getVidStreamOutput( udid )
    outSocket := vidConn.socket
    
    msgChan := make( chan ClientMsg )
    self.devTracker.addClient( udid, msgChan )
    
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
        
        select {
            case msg := <- msgChan:
                outSocket.WriteMessage( ws.TextMessage, []byte(msg.msg) )
                if msg.msgType == CMKick { break }
            default:       
        }
    }
    
    self.devTracker.deleteClient( udid )
    
    if conn != nil { conn.Close() }
    if outSocket != nil { outSocket.Close() }
}

func (self *ProviderHandler) handleProviderWS( c *gin.Context ) {
    s := self.sessionManager.GetSession( c )
    
    provider := self.sessionManager.session.Get( s, "provider" ).(ProviderOb)
    
    writer := c.Writer
    req := c.Request
    conn, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }

    provChan := make( chan ProvBase )
    provConn := NewProviderConnection( provChan )
    self.devTracker.setProvConn( provider.Id, provConn )
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
    
    self.devTracker.clearProvConn( provider.Id )
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

func (self *ProviderHandler) handleRegister( c *gin.Context ) {
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

func (self *ProviderHandler) handleProviderLogin( c *gin.Context ) {
    s := self.sessionManager.GetSession( c )
    
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
        
        self.sessionManager.session.Put( s, "provider", &ProviderOb{
            User: user,
            Id: provider.Id,
        } )
        self.sessionManager.WriteSession( c )
        
        c.Redirect( 302, "/provider/" )
        return
    } else {
        fmt.Printf("provider login failed [submit]%s != [db]%s\n", pass, provider.Password)
        c.Redirect( 302, "/provider/?fail=2" )
        return
    }
    
    self.showProviderLogin( c )
}

func (self *ProviderHandler) handleProviderLogout( c *gin.Context ) {
    s := self.sessionManager.GetSession( c )
    
    self.sessionManager.session.Remove( s, "provider" )
    self.sessionManager.WriteSession( c )
    
    c.Redirect( 302, "/" )
}

func (self *ProviderHandler) showProviderLogin( rCtx *gin.Context ) {
    rCtx.HTML( http.StatusOK, "providerLogin", gin.H{} )
}

func (self *ProviderHandler) showProviderRoot( c *gin.Context ) {
    c.HTML( http.StatusOK, "providerRoot", gin.H{} )
} 