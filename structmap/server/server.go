package server

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/goupdate/deadlock"

	"github.com/MasterDimmy/zipologger"

	"github.com/goupdate/compactmap/structmap"
	"github.com/valyala/fasthttp"
)

type Server[V any] struct {
	deadlock.RWMutex

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

	log := zipologger.NewLogger("./logs/server_api.log", 5, 5, 5, false)

	server := &Server[V]{
		storage:     storage,
		log:         log,
		storageName: storageName,
	}

	router := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		defer zipologger.HandlePanic()

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
		case "/api/updatecount":
			server.handleUpdateCount(ctx, storage)
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

func (s *Server[V]) SetLogger(log *zipologger.Logger) {
	s.log = log
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
		defer zipologger.HandlePanic()

		num := 0
		for range s.backupsTicker.C {
			fname := s.storageName + ".backup" + fmt.Sprintf("%d", num)
			s.log.Print("Autobackup to " + fname)
			err := s.storage.SaveAs(fname)
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
	if s.log == nil {
		return
	}
	switch s.logsLevel {
	case 0:
		return
	case 1:
		s.log.Printf("%s : %s", ctx.RemoteIP().String(), string(ctx.Request.URI().Path()))
	case 2:
		s.log.Printf("%s : %s [%s%s]", ctx.RemoteIP().String(), string(ctx.Request.URI().Path()), ctx.QueryArgs().String(), ctx.PostArgs().String())
	case 3:
		s.log.Printf("%s : %s [%s%s]", ctx.RemoteIP().String(), string(ctx.Request.URI().Path()), ctx.QueryArgs().String(), ctx.PostArgs().String())
		if len(response) > 0 {
			ret := ""
			for _, r := range response {
				rt := reflect.ValueOf(r)
				if rt.IsValid() && !rt.IsZero() && !rt.IsNil() {
					switch rt.Type().Kind() {
					case reflect.Slice:
						retv := ""
						for i := 0; i < rt.Len(); i++ {
							el := rt.Index(i)
							retv += fmt.Sprintf("%+v ", reflect.Indirect(el.Elem()))
						}
						ret += retv
					default:
						ret += fmt.Sprintf("%+v ", r)
					}
				}
			}
			s.log.Printf("%s", ret)
		}
	}
}

func (s *Server[V]) handleClear(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	storage.Clear()
	s.respondWithSuccess(ctx, nil)
}

func (s *Server[V]) handleAdd(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	var item V
	if err := json.Unmarshal(ctx.PostBody(), &item); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	id := storage.Add(&item)
	s.respondWithSuccess(ctx, map[string]int64{"id": id})
}

func (s *Server[V]) handleGet(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
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
	id, err := strconv.ParseInt(string(ctx.FormValue("id")), 10, 64)
	if err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	storage.Delete(id)
	s.respondWithSuccess(ctx, nil)
}

func (s *Server[V]) handleUpdate(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
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

func (s *Server[V]) handleUpdateCount(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
	var req struct {
		Random    bool                      `json:"random"` //update random count values?
		Count     int                       `json:"count"`  //how many elements to update
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
		Fields    map[string]interface{}    `json:"fields"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		s.respondWithError(ctx, err.Error())
		return
	}
	ids := storage.UpdateCount(req.Condition, req.Where, req.Fields, req.Count, req.Random)
	s.respondWithSuccess(ctx, map[string][]int64{"updated": ids})
}

func (s *Server[V]) handleSetField(ctx *fasthttp.RequestCtx, storage *structmap.StructMap[*V]) {
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
	var results = storage.GetAll()
	s.respondWithSuccess(ctx, results)
}
