package main

import (
    "errors"
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

func toJSON( s string ) template.JS {
    return template.JS( s )
}

func dictFunc(values ...interface{}) (map[string]interface{}, error) {
    if len(values)%2 != 0 {
        return nil, errors.New("invalid dict call")
    }
    dict := make(map[string]interface{}, len(values)/2)
    for i := 0; i < len(values); i+=2 {
        key, ok := values[i].(string)
        if !ok {
            return nil, errors.New("dict keys must be strings")
        }
        dict[key] = values[i+1]
    }
    return dict, nil
}

func tdefault(val interface{}, def interface{}) interface{} {
    switch val.(type) {
        case string:
            if len( val.(string) ) == 0 {
                fmt.Println("empty string")
                return def
            }
        case bool:
            if !val.(bool) {
                fmt.Println("false bool")
                return def
            }
        case nil:
            fmt.Println("nil type")
            return def
    }

    fmt.Println("fallthru")
    return val
}

func loadTemplatesFromDir( dir string ) (*template.Template, error) {
    t := template.New("")
    
    funcMap := template.FuncMap{
        "html": toHTML,
        "dict": dictFunc,
        "default": tdefault,
        "json": toJSON,
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