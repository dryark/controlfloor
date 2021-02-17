package main

import (
    "fmt"
    "os"
    _ "github.com/mattn/go-sqlite3"
    "xorm.io/xorm"
    uj "github.com/nanoscopic/ujsonin/mod"
)

var gDb *xorm.Engine

func openDbConnection() {
    gDb = openDb()
}

type DbDevice struct {
    Udid        string `xorm:"pk"`
    Name        string
    CustomName  string
    ProviderId  int64
    JsonInfo    string
    Width       int
    Height      int
    ClickWidth  int
    ClickHeight int
}
func (DbDevice) TableName() string {
    return "device"
}

type DbProvider struct {
    Id       int64
    Username string
    Password string
}
func (DbProvider) TableName() string {
    return "provider"
}

type DbConf struct {
    Id      int64
    RegPass string
}
func (DbConf) TableName() string {
    return "conf"
}

func openDb() ( *xorm.Engine ) {
    doesNotExist := false
    if _, err := os.Stat( "db.sqlite3" ); os.IsNotExist( err ) {
        doesNotExist = true
    }
    
    engine, err := xorm.NewEngine( "sqlite3", "./db.sqlite3" )
    if err != nil {
        panic( err )
    }
    
    if !doesNotExist {
        return engine
    }
            
    err = engine.Sync2( new( DbDevice ), new( DbProvider ), new( DbConf ) )
    if err != nil {
    }
    
    //addDummyDevice( engine, "4f5d", "Test" )
    addConf( engine, "doreg" )
    
    return engine
    
}

func getProvider( username string ) (*DbProvider) {
    var provider DbProvider
    has, err := gDb.Table(&provider).Where("username = ?", username).Get(&provider)
    if err != nil {
        return nil
    }
    if !has {
        return nil
    }
    return &provider
}

func getDevice( udid string ) (*DbDevice) {
    dev := DbDevice{
        Udid: udid,
    }
    has, err := gDb.Get( &dev )
    if err != nil {
        return nil
    }
    if !has {
        return nil
    }
    return &dev    
}

func addProvider( username string, password string ) (bool) {
    cur := getProvider( username )
    if cur != nil {
        fmt.Printf("Provider with username %s already existed\n", username )
        cur.Password = password
        fmt.Printf("  Updating password to %s\n", password)
        _, err := gDb.ID( cur.Id ).Update( cur )
        
        after := getProvider( username )
        fmt.Printf("  After update: %s\n", after.Password )
        
        if err != nil {
            panic( err )
        }
        return true
    }
    
    provider := DbProvider{
        Username: username,
        Password: password,
    }
    _, err := gDb.Insert( &provider )
    if err != nil {
        panic( err )
    }
    return false
}

func updateDeviceInfo( udid string, info string, pId int64 ) {
    root, _ := uj.Parse( []byte( info ) )    
    devNameNode := root.Get("DeviceName")
    if devNameNode == nil {
    }
    devName := devNameNode.String()
    
    dev := DbDevice{
        Udid: udid,
        Name: devName,
        JsonInfo: info,
        ProviderId: pId,
    }
    _, err := gDb.ID( udid ).Cols("JsonInfo", "Name", "ProviderId" ).Update( &dev )
    if err != nil {
        panic( err )
    }
}

func addConf( db *xorm.Engine, regPass string ) {
    conf := DbConf{
        RegPass: regPass,
    }
    _, err := db.Insert( &conf )
    if err != nil {
        panic( err )
    }
}

func addDevice( udid string, name string, pId int64, width int, height int, clickWidth int, clickHeight int ) {
    fmt.Printf("Adding device:\n"+
        "  udid:%s\n"+
        "  name:%s\n"+
        "  clickWidth:%d\n"+
        "  clickHegiht:%d\n",
        udid,name,clickWidth,clickHeight)
    dev := DbDevice{
        Udid: udid,
        Name: name,
        ProviderId: pId,
        Width: width,
        Height: height,
        ClickWidth: clickWidth,
        ClickHeight: clickHeight,
    }
    cur := getDevice( udid )
    if cur != nil {
        fmt.Printf("Device with udid %s already existed\n", dev.Udid )
        _, err := gDb.ID( udid ).Update( &dev ) // Cols("Name","ProviderId","ClickWidth","ClickHeight").
        if err != nil {
            panic( err )
        }
        return
    }
    
    _, err := gDb.Insert( &dev )
    if err != nil {
        panic( err )
    }
}

func addDummyDevice( db *xorm.Engine, udid string, name string ) {
    device := DbDevice{
        Udid: udid,
        Name: name,
    }
    _, err := db.Insert( &device )
    if err != nil {
        panic( err )
    }
}

func getConf() (*DbConf) {
    var confs [] DbConf
    err := gDb.Find( &confs )
    if err != nil {
        panic( err )
    }
    
    return &confs[0]
}