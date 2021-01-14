package main

import (
    "crypto/rand"
    "fmt"
    "encoding/hex"
    "net/http"
    
    //ecies "github.com/ecies/go"
    "github.com/gorilla/websocket"
    //uj "github.com/nanoscopic/ujsonin/mod"
    "github.com/gin-gonic/gin"
)

func registerProviderRoutes( r *gin.Engine ) (*gin.RouterGroup) {
    r.POST("/register", handleRegister )
    r.GET("/provider/login", showProviderLogin )
    r.GET("/provider/logout", handleProviderLogout )
    r.POST("/provider/login", handleProviderLogin )
    
    pAuth := r.Group("/provider")
    pAuth.Use( NeedProviderAuth() )
    pAuth.GET("/", showProviderRoot )
    pAuth.GET("/ws", handleProviderWS )
    
    return pAuth
}

func NeedProviderAuth() gin.HandlerFunc {
    return func( c *gin.Context ) {
        sCtx := getSession( c )
        
        loginI := session.Get( sCtx, "provider" )
        
        if loginI == nil {
            c.Redirect( 302, "/provider/login" )
            c.Abort()
            
            return
        }
        
        c.Next()
    }
}

var wsupgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func handleProviderWS( ginCtx *gin.Context ) {
    writer := ginCtx.Writer
    req := ginCtx.Request
    ws, err := wsupgrader.Upgrade( writer, req, nil )
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }

    /*mType, introText, err := ws.ReadMessage()
    if err != nil {
        fmt.Println("Error reading ws intro: %+v", err)
        return
    }
    intro, _ := uj.Parse( introText )
    idNode := intro.Get("id")
    if idNode == nil {
        fmt.Println("Ws intro does not contain id")
        return
    }
    id := idNode.String()
    challenge := rangHex()*/
    // encrypt the challenge and send it
    // wait for the challenge to be send back decrypted
    
    
    for {
        t, msg, err := ws.ReadMessage()
        if err != nil {
            break
        }
        ws.WriteMessage(t, msg)
    }
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

func handleProviderLogin( c *gin.Context ) {
    s := getSession( c )
    
    user := c.PostForm("user")
    pass := c.PostForm("pass")
    
    // ensure the user is legit
    provider := getProvider( user )
    if provider == nil {
        c.Redirect( 302, "/provider/?fail=1" )
        return
    }
    
    if pass == provider.Password {
        fmt.Printf("login ok\n")
        
        session.Put( s, "provider", "test" )
        writeSession( c )
        
        c.Redirect( 302, "/provider/" )
        return
    } else {
        fmt.Printf("provider login failed\n")
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