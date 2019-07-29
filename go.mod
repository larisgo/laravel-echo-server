module github.com/larisgo/laravel-echo-server

go 1.12

replace (
	github.com/pschlump/socketio => github.com/zishang520/socketio v2.0.1+incompatible
	github.com/tcnksm/go-input => github.com/zishang520/go-input v1.0.0
)

require (
	github.com/go-redis/redis v6.15.3+incompatible
	github.com/gookit/color v1.1.7
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/pschlump/json v1.12.0 // indirect
	github.com/pschlump/socketio v0.0.0-00010101000000-000000000000
	github.com/smartystreets/goconvey v0.0.0-20190710185942-9d28bd7c0945 // indirect
	github.com/tcnksm/go-input v0.0.0-00010101000000-000000000000
)
