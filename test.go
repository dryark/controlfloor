package main

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
)

type TestHandler struct {
    r              *gin.Engine
    sessionManager *cfSessionManager
}

func NewTestHandler(
    r              *gin.Engine,
    sessionManager *cfSessionManager,
) *TestHandler {
    return &TestHandler{
        r,
        sessionManager,
    }
}

func (self *TestHandler) registerTestRoutes() {
    self.r.GET("/test", self.handleTest )
    //self.r.GET( "/test",  )
}

func (self *TestHandler) handleTest( c *gin.Context ) {
    s := self.sessionManager.GetSession( c )
    
    var num int
    prevVal := self.sessionManager.session.GetString( s, "test" )
    //fmt.Printf("prevval:%s\n", prevVal )
    if prevVal == "" {
        num = 0
    } else {
        num, _ = strconv.Atoi( prevVal )
        num = num + 1
    }
    
    self.sessionManager.session.Put( s, "test", strconv.Itoa( num ) )
    
    self.sessionManager.WriteSession( c )
    
    //c.Header("X-Session", token )
    //c.Header("X-Session-Expiry", expiry.Format( http.TimeFormat ) )
     
    c.Data( http.StatusOK, "text/html", []byte("test - num:" + strconv.Itoa( num )) )
}