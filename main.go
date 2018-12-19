package main

import (
	"github.com/anakin/mock/dbops"
	"github.com/anakin/mock/mocker"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
)


func RegisterHandlers() *httprouter.Router {
	router := httprouter.New()
	router.GET("/mock/:id", MockHandler)
	return router
}

func MockHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	mockId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		log.Println("param error", err.Error())
	}
	mock, err := dbops.GetMockById(mockId)
	if err != nil {
		log.Println("get mock error", err.Error())
	}
	m, err := mocker.InitMocker(mock)
	m.UpStatus(1, float64(0))
	if err != nil {
		log.Println("init mock error:", err)
	}
	go m.Loop()
}

func DoMock() {
	mock, err := dbops.GetMocks()
	if err != nil {
		log.Println("get mock error", err.Error())
	}
	m, err := mocker.InitMocker(mock)
	m.UpStatus(1, float64(0))
	if err != nil {
		log.Println("init mock error:", err)
	}
	m.Loop()
}
func main() {
	//DoMock()
	_ = http.ListenAndServe(":8080", RegisterHandlers())
}
