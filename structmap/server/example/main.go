package main

import (
	"github.com/MasterDimmy/go-ctrlc"
	"github.com/MasterDimmy/zipologger"
	"github.com/goupdate/compactmap/structmap/server"
)

type some1 struct {
	Id   int64
	Aba  string
	Haba int
}

func main() {
	defer zipologger.Wait()
	defer zipologger.HandlePanic()

	var ctrl ctrlc.CtrlC
	log := zipologger.NewLogger("./logs/server.log", 1, 1, 1, false)

	srv, err := server.New[some1]("storage1")
	if err != nil {
		log.Print(err.Error())
		return
	}

	if srv == nil {
		log.Print("no srv!")
		return
	}

	srv.SetLoggingLevel(2)

	defer srv.Shutdown()
	defer ctrl.DeferThisToWaitCtrlC()

	go func() {
		srv := srv.GetFasthttpServer()
		if srv != nil {
			err := srv.ListenAndServe(":80")
			if err != nil {
				log.Print("server error: " + err.Error())
			}
		} else {
			log.Print("no srv server!")
		}
		ctrl.ForceStopProgram()
	}()

	ctrl.InterceptKill(true, func() {
		log.Println("software was stopped via Ctrl+C")
	})
}
