package v1

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/teaelephant/TeaElephantMemory/common"
)

var errorEmptyID = errors.New("empty id")

type Storage interface {
	WriteRecord(rec *common.TeaData) (record *common.Tea, err error)
	ReadRecord(id string) (record *common.Tea, err error)
	ReadAllRecords(search string) ([]common.Tea, error)
	Update(id string, rec *common.TeaData) (record *common.Tea, err error)
	Delete(id string) error
}

type errorCreator interface {
	ResponseError(w http.ResponseWriter, err common.Error)
}

type transport interface {
	Response(w http.ResponseWriter, answer interface{}) error
}

type RecordManager struct {
	Storage
	errorCreator
	transport
}

// Create new record in Storage
func (m *RecordManager) NewRecord(w http.ResponseWriter, r *http.Request) {
	logrus.Info("new record")
	record := new(common.TeaData)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("read request httperror")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: err})
		return
	}
	if err := json.Unmarshal(data, record); err != nil {
		logrus.WithError(err).Error("unmarshal request httperror")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: err})
		return
	}
	recWithID, err := m.Storage.WriteRecord(record)
	if err != nil {
		logrus.WithError(err).Error("write request httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
	if err := m.transport.Response(w, recWithID); err != nil {
		logrus.WithError(err).Error("write response httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
}

// Read record from Storage by id
func (m *RecordManager) ReadRecord(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	logrus.WithField("id", id).Info("read record")
	if id == "" {
		logrus.Error("empty id")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: errorEmptyID})
		return
	}
	rec, err := m.Storage.ReadRecord(id)
	if err != nil {
		logrus.WithError(err).Error("read from Storage httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
	if err := m.transport.Response(w, rec); err != nil {
		logrus.WithError(err).Error("write response httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
}

func (m *RecordManager) ReadAllRecords(w http.ResponseWriter, r *http.Request) {
	logrus.Info("read record")
	name := r.URL.Query().Get("name")
	logrus.WithField("name", name).Info("search record by name")
	rec, err := m.Storage.ReadAllRecords(name)
	if err != nil {
		logrus.WithError(err).Error("read from Storage httperror")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: err})
		return
	}
	if err := m.transport.Response(w, rec); err != nil {
		logrus.WithError(err).Error("write response httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
}

func (m *RecordManager) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	logrus.WithField("id", id).Info("update record")
	if id == "" {
		logrus.Error("empty id")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: errorEmptyID})
		return
	}
	record := new(common.TeaData)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("read request httperror")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: err})
		return
	}
	if err := json.Unmarshal(data, record); err != nil {
		logrus.WithError(err).Error("unmarshal request httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
	rec, err := m.Storage.Update(id, record)
	if err != nil {
		logrus.WithError(err).Error("read from Storage httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
	if err := m.transport.Response(w, rec); err != nil {
		logrus.WithError(err).Error("write response httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
}

func (m *RecordManager) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	logrus.WithField("id", id).Info("delete record")
	if id == "" {
		logrus.Error("empty id")
		m.ResponseError(w, common.Error{Code: http.StatusBadRequest, Msg: errorEmptyID})
		return
	}
	if err := m.Storage.Delete(id); err != nil {
		logrus.WithError(err).Error("delete from Storage httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
	}
	if err := m.transport.Response(w, struct {
		ID string `json:"id"`
	}{ID: id}); err != nil {
		logrus.WithError(err).Error("write response httperror")
		m.ResponseError(w, common.Error{Code: http.StatusInternalServerError, Msg: err})
		return
	}
}

func New(s Storage, errorCreator errorCreator, tr transport) *RecordManager {
	return &RecordManager{
		Storage:      s,
		errorCreator: errorCreator,
		transport:    tr,
	}
}
