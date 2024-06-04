package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/MasterDimmy/zipologger"

	"github.com/goupdate/compactmap/structmap"
	"github.com/valyala/fasthttp"
)

type Server[V any] struct {
	sync.RWMutex

	storage     *structmap.StructMap[*V]
	srv         *fasthttp.Server
	storageName string

	backupsTicker *time.Ticker

	logsLevel int //0 = OFF, 1=CALLS, 2=CALLS+DATA
	log       *zipologger.Logger
}

func (s *Server[V]) Shutdown() {
	if s.srv != nil {
		s.srv.Shutdown()
	}
	if s.storage != nil {
		s.storage.Save()
	}
}

func New[V any](storageName string) (*Server[V], error) {
	var err error

	storage, err := structmap.New[*V](storageName, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize storage: %v", err)
	}

	log := zipologger.NewLogger("./logs/structmap_server.log", 5, 5, 5, false)

	server := &Server[V]{
		storage:     storage,
		log:         log,
		storageName: storageName,
	}

	router := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/api/clear":
			server.handleClear(ctx, storage)
		case "/api/add":
			server.handleAdd(ctx, storage)
		case "/api/get":
			server.handleGet(ctx, storage)
		case "/api/delete":
			server.handleDelete(ctx, storage)
		case "/api/update":
			server.handleUpdate(ctx, storage)
		case "/api/setfield":
			server.handleSetField(ctx, storage)
		case "/api/setfields":
			server.handleSetFields(ctx, storage)
		case "/api/find":
			server.handleFind(ctx, storage)
		case "/api/all":
			server.handleAll(ctx, storage)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	})

	server.srv = &fasthttp.Server{Handler: router}
	return server, nil
}

func (s *Server[V]) GetFasthttpServer() *fasthttp.Server {
	return s.srv
}

func (s *Server[V]) GetStorage() *structmap.StructMap[*V] {
	return s.storage
}

func (s *Server[V]) EnableBackupsEvery(interval time.Duration, storeBackups int) {
	s.Lock()
	defer s.Unlock()

	if s.backupsTicker != nil {
		s.backupsTicker.Stop()
	}
	s.backupsTicker = time.NewTicker(interval)
	go func() {
		num := 0
		for _ = range s.backupsTicker.C {
			err := s.storage.SaveAs(s.storageName + ".backup" + fmt.Sprintf("%d", num))
			if err != nil {
				s.log.Printf("ERROR: %v", err.Error())
			}
			num++
			num = num % storeBackups
		}

	}()
}

func (s *Server[V]) respondWithError(ctx *fasthttp.RequestCtx, message string) {
	s.logAction(ctx, []string{"ERROR", message})

	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetBodyString(message)
}

func (s *Server[V]) respondWithSuccess(ctx *fasthttp.RequestCtx, response interface{}) {
	s.logAction(ctx, response)

	ctx.SetStatusCode(fasthttp.StatusOK)
	if response != nil {
		body, _ := json.Marshal(response)
		ctx.SetBody(body)
	}
}

// l = 0=OFF,1=URL,2=DATA
func (s *Server[V]) SetLoggingLevel(l int) {
	s.logsLevel = l
}

func (s *Server[V]) logAction(ctx *fasthttp.RequestCtx, response ...interface{}) {
	switch s.logsLevel {
	case 0:
		return
	case 1:
		s.log.Printf("%s : %s", ctx.RemoteIP().String(), string(ctx.Request.URI().Path()))
	case 2:
		s.log.Printf("%s : %s", ctx.RemoteIP().String(), string(ctx.Request.URI().String()))
		if len(response) > 0 {
			ret := ""
			for _, r := range response {
				ret += fmt.Sprintf("%+v ", r)
			}
			s.log.Printf("%s", ret)
		}
	}
}

func (s *Server[V]) handleClear(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	storage.Clear()
	s.respondWithSuccess(ctx, nil)
}

func (s *Server[V]) handleAdd(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	var item V
	if err := json.Unmarshal(ctx.PostBody(), &item); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	id := storage.Add(&item)
	s.respondWithSuccess(ctx, map[string]int64{"id": id})
}

func (s *Server[V]) handleGet(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	id, err := strconv.ParseInt(string(ctx.FormValue("id")), 10, 64)
	if err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	item, found := storage.Get(id)
	if !found {
		s.respondWithSuccess(ctx, nil)
		return
	}
	s.respondWithSuccess(ctx, item)
}

func (s *Server[V]) handleDelete(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	id, err := strconv.ParseInt(string(ctx.FormValue("id")), 10, 64)
	if err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	storage.Delete(id)
	s.respondWithSuccess(ctx, nil)
}

func (s *Server[V]) handleUpdate(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	var req struct {
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
		Fields    map[string]interface{}    `json:"fields"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	count := storage.Update(req.Condition, req.Where, req.Fields)
	s.respondWithSuccess(ctx, map[string]int{"updated": count})
}

func (s *Server[V]) handleSetField(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	var req struct {
		Id    int64       `json:"id"`
		Field string      `json:"field"`
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	if !storage.SetField(req.Id, req.Field, req.Value) {
		s.respondWithError(ctx, "Failed to set field")
		return
	}
	s.respondWithSuccess(ctx, nil)
}

func (s *Server[V]) handleSetFields(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	var req struct {
		Id     int64                  `json:"id"`
		Fields map[string]interface{} `json:"fields"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	if !storage.SetFields(req.Id, req.Fields) {
		s.respondWithError(ctx, "Failed to set fields")
		return
	}
	s.respondWithSuccess(ctx, nil)
}

func (s *Server[V]) handleFind(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	var req struct {
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	results := storage.Find(req.Condition, req.Where...)
	s.respondWithSuccess(ctx, results)
}

func (s *Server[V]) handleAll(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	s.logAction(ctx)

	var results = storage.GetAll()
	s.respondWithSuccess(ctx, results)
}
