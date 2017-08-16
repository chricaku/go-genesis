// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package stopdaemons

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/system"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/*
#include <stdio.h>
#include <signal.h>

extern void go_callback_int();
static inline void SigBreak_Handler(int n_signal){
    printf("closed\n");
	go_callback_int();
}
static inline void waitSig() {
    #if (WIN32 || WIN64)
    signal(SIGBREAK, &SigBreak_Handler);
    signal(SIGINT, &SigBreak_Handler);
    #endif
}
*/
import (
	"C"
)

//export go_callback_int
func go_callback_int() {
	SigChan <- syscall.Signal(1)
}

// SigChan is a channel
var SigChan chan os.Signal

func waitSig() {
	C.waitSig()
}

// Signals waits for Interrupt os.Kill signals
func WaintForSignals() {
	SigChan = make(chan os.Signal, 1)
	waitSig()
	var Term os.Signal = syscall.SIGTERM
	go func() {
		signal.Notify(SigChan, os.Interrupt, os.Kill, Term)
		<-SigChan
		fmt.Println("got kill signal")

		if utils.CancelFunc != nil {
			utils.CancelFunc()
			for i := 0; i < utils.DaemonsCount; i++ {
				name := <-utils.ReturnCh
				log.Debugf("daemon %s stopped", name)
			}

			log.Debug("Daemons killed")
			fmt.Println("Daemons killed")
		}

		system.FinishThrust()

		if model.DBConn != nil {
			err := model.GormClose()
			if err != nil {
				log.Error("gorm close error: %s", utils.ErrInfo(err).Error())
			}
		}

		err := os.Remove(*utils.Dir + "/daylight.pid")
		if err != nil {
			log.Error("can't remove pid file: %s", err)
		} else {
			fmt.Println("removed " + *utils.Dir + "/daylight.pid")
		}

		os.Exit(1)

	}()
}
