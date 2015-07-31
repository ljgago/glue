/*
 *  Glue - Robust Go and Javascript Socket Library
 *  Copyright DesertBit
 *  Author: Roland Singer
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package glue

//####################//
//### Read Handler ###//
//####################//

type readHandler struct {
	Stopped chan struct{}
}

func (h *readHandler) Stop() {
	// Signal the stop request by closing the channel.
	close(h.Stopped)
}

//#################################//
//### Additional Socket Methods ###//
//#################################//

func (s *Socket) newReadHandler() *readHandler {
	// Lock the mutex.
	s.readHandlerMutex.Lock()
	defer s.readHandlerMutex.Unlock()

	// If a previous handler is set, then stop it first.
	if s.readHandler != nil {
		s.readHandler.Stop()
	}

	// Create a new read handler.
	s.readHandler = &readHandler{
		Stopped: make(chan struct{}),
	}

	return s.readHandler
}

func (s *Socket) stopReadHandler() {
	// Lock the mutex.
	s.readHandlerMutex.Lock()
	defer s.readHandlerMutex.Unlock()

	// Stop the read handler if set.
	if s.readHandler != nil {
		s.readHandler.Stop()
		s.readHandler = nil
	}
}

//##########################//
//### Broadcast Handlers ###//
//##########################//

type Router struct {
  sockets		     	map [*Socket]bool    // Registered clients.
  broadcast  			chan []byte         // Inbound messages from the clients.
  register    		chan *Socket        // Register requests from the clients.
  unregister  		chan *Socket       	// Unregister requests from clients.

}

var Server = Router {
  sockets: 		   	make(map [*Socket]bool),
  broadcast:  		make(chan []byte),
  register:   		make(chan *Socket),
  unregister: 		make(chan *Socket),
}

func (r *Router) runRouter() {
  go func() {
  	for {
	    select {
	    case s := <-r.register:
	      r.sockets[s] = true
	    case s := <-r.unregister:
	      if _, ok := r.sockets[s]; ok {
	        delete(r.sockets, s)
	      }
	    case m := <-r.broadcas:
	      for s := range r.sockets {
	        s.Write(m)
          /*select {
	        case s.writeChan <- m:
	        default:
	          delete(r.sockets, s)
	        }*/
	      }
	    }
	  }
  }()
}

func (r *Router) writeBroadcast(rawData string) {
  r.broadcast <- []byte(rawData)
}

func WriteBroadcast(event string, msg string) {
	rawData := `{"event":` + event + `,"data":` + msg + `}`
	writeBroadcast(rawData)
}

func RunRouter() {
  runRouter()
}