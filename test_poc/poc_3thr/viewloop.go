package poc_3thr

//
//import (
//	"context"
//	"fmt"
//	"sync"
//)
//
//type View struct {
//	currentView
//}
//
//func (view *View) startView(parentCtx context.Context, wg *sync.WaitGroup) {
//
//	Log("H=%d view.startView() starting *VIEWLOOP* goroutine")
//	id := parentCtx.Value("ID")
//	newID := fmt.Sprintf("%s|V=%d", id, view.currentView)
//	// TODO Do something with cancel func?
//	mainLoopCtx, cancel := context.WithCancel(context.WithValue(parentCtx, "ID", newID))
//	term.cancel = cancel
//
//	go term.TermLoop(mainLoopCtx, wg)
//}
