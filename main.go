package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    openDbConnection()

    r := gin.New()
    initTemplates( r )
    r.Static("/assets", "./assets")
    initSessionManager( r )
    uAuth := registerUserRoutes( r )
    pAuth := registerProviderRoutes( r )
    registerDeviceRoutes( r, pAuth, uAuth )
    registerTestRoutes( r )
    
    r.Run( ":8080" )
}