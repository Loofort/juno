package ckeck

import (
	"juno/common/io"
	"log"
	"net/http"
)

// initialize custom loger, because we will use log.Output(calldepth, msg)
var clog = log.New(os.Stderr, "", log.LstdFlags)

func DBErr(w http.ResponseWriter, err error) bool {
	if err != nil {
		// send to user common err message
		io.ErrServer(w, io.ERR_DB)

		// log real db error
		clog.Output(2, err.Error())

		return true
	}
	return false
}

func InputErr(r *http.Request, obj interface{}) bool {
	if err := io.Input(r, obj); err != nil {
		io.ErrClient(w, io.ERR_REQ)
		return true
	}
	return false
}
