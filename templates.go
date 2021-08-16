package main

import (
    "errors"
    "fmt"
    "html/template"
    "github.com/gin-gonic/gin"
    "github.com/foolin/goview"
    "github.com/foolin/goview/supports/ginview"
)

func initTemplates( r *gin.Engine, config *Config ) {
    r.HTMLRender = ginview.New( goview.Config{
        Root:         fmt.Sprintf( "tmpl/%s", config.theme ),
        Extension:    ".tmpl",
        Partials:     []string{"sidebar"},
        Funcs:        createFuncMap(),
        DisableCache: config.disableCache,
    } )
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

func createFuncMap() template.FuncMap {
    return template.FuncMap{
        "html":    toHTML,
        "dict":    dictFunc,
        "default": tdefault,
        "json":    toJSON,
    }
}