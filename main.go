package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "os"
    "os/exec"
    uc "github.com/nanoscopic/uclop/mod"
    cfauth "github.com/nanoscopic/controlfloor_auth"
)

func main() {
    uclop := uc.NewUclop()
    uclop.AddCmd( "run", "Run ControlFloor", runMain, nil )
    uclop.AddCmd( "devs", "List registered devices", runListDevs, nil )
    uclop.AddCmd( "prov", "List providers", runListProv, nil )
    uclop.AddCmd( "conf", "Dump configuration", runDumpConf, nil )
    uclop.Run()
}

func runDumpConf( *uc.Cmd ) {
    conf := NewConfig( "config.json", "default.json" )
    fmt.Printf("%s\n", conf )
}

func runListDevs( *uc.Cmd ) {
    openDbConnection()
    
    var devices [] DbDevice
    err := gDb.Find( &devices )
    if err != nil {
        panic( err )
    }
    
    for _, device := range devices {
        fmt.Printf("Name: %s\nUdid: %s\nProvider Id: %d\n\n",
            device.Name, device.Udid, device.ProviderId )
    }
}

func runListProv( *uc.Cmd ) {
    openDbConnection()
    
    var provs [] DbProvider
    err := gDb.Find( &provs )
    if err != nil {
        panic( err )
    }
    
    for _, prov := range provs {
        fmt.Printf("Username: %s\nProvider Id: %d\n\n",
            prov.Username, prov.Id )
    }
}

func runMain( *uc.Cmd ) {
    conf := NewConfig( "config.json", "default.json" )
    
    openDbConnection()

    gin.SetMode(gin.ReleaseMode)
    r := gin.New()
            
    initTemplates( r )
    r.Static("/assets", "./assets")
    sessionManager := NewSessionManager( r )
    
    devTracker := NewDevTracker()
    
    var authHandler cfauth.AuthHandler
    if conf.auth == "mod" {
        authHandler = cfauth.NewAuthHandler( conf.root, sessionManager )
    }
    
    uh := NewUserHandler( authHandler, r, devTracker, sessionManager )
    uAuth := uh.registerUserRoutes()
    
    ph := NewProviderHandler( r, devTracker, sessionManager )
    pAuth := ph.registerProviderRoutes()
    
    dh := NewDevHandler( pAuth, uAuth, devTracker, sessionManager )
    dh.registerDeviceRoutes()
    
    th := NewTestHandler( r, sessionManager )
    th.registerTestRoutes()
    
    var err error
    protocol := "http"
    if conf.https {
        protocol = "https"
        if conf.crt == "server.crt" && !fileExists("server.crt") {
            gen_cert()
        }
        err = http.ListenAndServeTLS( conf.listen, conf.crt, conf.key, r )
    } else {
    	err = http.ListenAndServe( conf.listen, r )
    }
    fmt.Printf("%s ListenAndServe Error %s\n", protocol, err)
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func gen_cert() {
    out, err := exec.Command( "/usr/bin/perl", "gencert.pl" ).Output()
    if err != nil {
        fmt.Printf("Error from cert gen: %s\n", err )
        return
    }
    fmt.Println( out )
}