module main.go

go 1.12

//replace github.com/nanoscopic/controlfloor_auth => ../controlfloor_auth
replace github.com/nanoscopic/controlfloor/docs => ./docs

//replace github.com/nanoscopic/ujsonin/v2 => ../ujsonin/v2

require (
	github.com/alexedwards/scs/v2 v2.4.0
	github.com/foolin/goview v0.3.0
	github.com/gin-gonic/gin v1.7.0
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/nanoscopic/controlfloor/docs v0.0.0-00010101000000-000000000000
	github.com/nanoscopic/controlfloor_auth v1.1.0
	github.com/nanoscopic/uclop v1.1.0
	github.com/nanoscopic/ujsonin/v2 v2.0.6
	github.com/sirupsen/logrus v1.8.0
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2
	github.com/swaggo/gin-swagger v1.3.1
	github.com/swaggo/swag v1.7.1 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.5 // indirect
	xorm.io/xorm v1.0.6
)
