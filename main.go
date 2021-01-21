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
    
    devTracker := NewDevTracker()
    
    uAuth := registerUserRoutes( r, devTracker )
    pAuth := registerProviderRoutes( r, devTracker )
    registerDeviceRoutes( r, pAuth, uAuth, devTracker )
    registerTestRoutes( r )
    
    r.Run( ":8080" )
}