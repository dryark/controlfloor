package main

import (
    "fmt"
    "html/template"
    "io/ioutil"
    "strings"
    "github.com/gin-gonic/gin"
)

func initTemplates( r *gin.Engine ) {
    templates, err := loadTemplatesFromDir( "tmpl" )
    if err != nil {
        panic( err )
    }
    r.SetHTMLTemplate( templates )
}

func toHTML( s string ) template.HTML {
    return template.HTML( s )
}

func loadTemplatesFromDir( dir string ) (*template.Template, error) {
    t := template.New("")
    
    funcMap := template.FuncMap{
        "html": toHTML,
    }
    t = t.Funcs( funcMap )
    
    files, err := ioutil.ReadDir( dir )
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        name := file.Name()
        if file.IsDir() || !strings.HasSuffix( name, ".tmpl" ) {
            continue
        }
        fullName := dir + "/" + name
        fmt.Printf("Loading template from %s\n", name )
        
        content, err := ioutil.ReadFile( fullName )
        if err != nil {
            return nil, err
        }
        refName := name[ 0 : len( name ) - 5 ]
        
        t, err = t.New( refName ).Parse( string( content ) )
        if err != nil {
            return nil, err
        }
    }
        
    return t, nil
}