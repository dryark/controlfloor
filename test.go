package main

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
)

func registerTestRoutes( r *gin.Engine ) {
    r.GET("/test", handleTest )
    //r.GET( "/test",  )
}

func handleTest( c *gin.Context ) {
    s := getSession( c )
    
    var num int
    prevVal := session.GetString( s, "test" )
    //fmt.Printf("prevval:%s\n", prevVal )
    if prevVal == "" {
        num = 0
    } else {
        num, _ = strconv.Atoi( prevVal )
        num = num + 1
    }
    
    session.Put( s, "test", strconv.Itoa( num ) )
    
    writeSession( c )
    
    //c.Header("X-Session", token )
    //c.Header("X-Session-Expiry", expiry.Format( http.TimeFormat ) )
     
    c.Data( http.StatusOK, "text/html", []byte("test - num:" + strconv.Itoa( num )) )
}