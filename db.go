package main

import (
    "fmt"
    "os"
    _ "github.com/mattn/go-sqlite3"
    "xorm.io/xorm"
)

var gDb *xorm.Engine

func openDbConnection() {
    gDb = openDb()
}

type DbDevice struct {
    Id int64
    Udid string
    Name string
    CustomName string
    ProviderId int64
}
func (DbDevice) TableName() string {
    return "device"
}

type DbProvider struct {
    Id int64
    Username string
    Password string
}
func (DbProvider) TableName() string {
    return "provider"
}

type DbConf struct {
    Id int64
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
    
    /*
    fmt.Printf("Creating sqlite3 database\n")
    
    db, err := sql.Open( "sqlite3", "./db.sqlite3" )
    if err != nil {
        panic( err )
    }
    
    doCreate := `
        create table devices (
            id integer not null primary key,
            udid text,
            name text,
            customName text
        )
    `
    
    _, err = db.Exec( doCreate )
    if err != nil {
        panic( err )
    }
    
    db.Close()
    */
        
    err = engine.Sync2( new( DbDevice ), new( DbProvider ), new( DbConf ) )
    if err != nil {
    }
    
    addDummyDevice( engine, "4f5d", "Test" )
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

func addProvider( username string, password string ) (bool) {
    cur := getProvider( username )
    if cur != nil {
        fmt.Printf("Provider with username %s already existed\n", username )
        cur.Password = password
        _, err := gDb.ID(cur.Id).Update( cur )
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

func addConf( db *xorm.Engine, regPass string ) {
    conf := DbConf{
        RegPass: regPass,
    }
    _, err := db.Insert( &conf )
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
