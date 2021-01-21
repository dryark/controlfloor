package main

import (
    "context"
    "fmt"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/alexedwards/scs/v2"
    "encoding/gob"
)

func init() {
	gob.Register(ProviderOb{})
}

var session *scs.SessionManager

func initSessionManager( r *gin.Engine ) {
    session = scs.New()
    
    r.Use( Sessions() )
    //db, _ := sql.Open( "sqlite3", "sessions.db" )
    //session.Store = sqlite3store.New( db )
}

func Sessions() gin.HandlerFunc {
    return func( c *gin.Context ) {
        //fmt.Printf("Sessions")
        
        r := c.Request
        
        token, _ := c.Cookie( "session" )
        
        ctx, _ := session.Load( r.Context(), token )
        if ctx == nil {
            fmt.Println("no session")
        } else {
            c.Set("session", ctx)
        }
        //fmt.Printf("token:%s\n", token )
        
        c.Next()
    }
}

func getSession( rCtx *gin.Context ) ( context.Context ) {
    ctx, _ := rCtx.Get("session")
    ctx2 := ctx.(context.Context)
    return ctx2
}

func writeSession( c *gin.Context ) {
    sI, _ := c.Get("session")
    
    s := sI.(context.Context)
    
    status := session.Status( s )
        
    if status == scs.Unmodified {
        return
    }
    
    //var cExpires int
    var cMaxAge int
    var token string
    var expiry time.Time
    
    switch status {
    case scs.Modified:
        token, expiry, _ = session.Commit( s )
        //fmt.Println("session committed")
        //cExpires = time.Unix( expiry.Unix() + 1, 0 )
        cMaxAge  = int( time.Until( expiry ).Seconds() + 1 )
    
    case scs.Destroyed:
        //cExpires = time.Unix( 1, 0 )
        cMaxAge = -1
    }
    
    c.SetCookie(
        "session",
        token,
        cMaxAge,
        "/",
        session.Cookie.Domain,
        session.Cookie.Secure,
        session.Cookie.HttpOnly )
}