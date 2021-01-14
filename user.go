package main

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
)

func registerUserRoutes( r *gin.Engine ) (*gin.RouterGroup) {
    r.GET("/login", showUserLogin )
    r.GET("/logout", handleUserLogout )
    r.POST("/login", handleUserLogin )
    userAuth := r.Group("/")
    userAuth.Use( NeedUserAuth() )
    userAuth.GET("/", showUserRoot )
    return userAuth
}

func NeedUserAuth() gin.HandlerFunc {
    return func( c *gin.Context ) {
        sCtx := getSession( c )
        
        loginI := session.Get( sCtx, "user" )
        
        if loginI == nil {
            c.Redirect( 302, "/login" )
            c.Abort()
            
            return
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
        output = output + fmt.Sprintf("Name: %s<br>\n", device.Name )
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