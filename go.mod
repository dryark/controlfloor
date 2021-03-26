module main.go

go 1.12

//replace github.com/nanoscopic/controlfloor_auth => ../controlfloor_auth

require (
	github.com/alexedwards/scs/sqlite3store v0.0.0-20201122155747-696f8e8a5fe2
	github.com/alexedwards/scs/v2 v2.4.0
	github.com/ecies/go v1.0.1
	github.com/gin-gonic/gin v1.6.3
	github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/nanoscopic/controlfloor_auth v1.0.1
	github.com/nanoscopic/uclop v1.1.0
	github.com/nanoscopic/ujsonin v1.13.0
	github.com/nanoscopic/ujsonin/v2 v2.0.4
	github.com/sirupsen/logrus v1.8.0
	xorm.io/xorm v1.0.6
)
