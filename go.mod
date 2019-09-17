module github.com/larisgo/laravel-echo-server

go 1.13

replace (
	github.com/pschlump/socketio => github.com/zishang520/socketio v2.0.3+incompatible
	github.com/tcnksm/go-input => github.com/zishang520/go-input v1.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-redis/redis v6.15.5+incompatible
	github.com/gookit/color v1.2.0
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pschlump/json v1.12.0 // indirect
	github.com/pschlump/socketio v0.0.0-00010101000000-000000000000
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 // indirect
	golang.org/x/sys v0.0.0-20190412213103-97732733099d // indirect
)
