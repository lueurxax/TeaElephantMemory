package tea_manager

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/teaelephant/TeaElephantMemory/common"
	"github.com/teaelephant/TeaElephantMemory/internal/tea_manager/subscribers"
	gqlCommon "github.com/teaelephant/TeaElephantMemory/pkg/api/v2/common"
	model "github.com/teaelephant/TeaElephantMemory/pkg/api/v2/models"
)

type TeaManager interface {
	Create(ctx context.Context, data *common.TeaData) (tea *common.Tea, err error)
	Update(ctx context.Context, id uuid.UUID, rec *common.TeaData) (record *common.Tea, err error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (record *common.Tea, err error)
	List(ctx context.Context, search *string) ([]common.Tea, error)
	SubscribeOnCreate(ctx context.Context) (<-chan *model.Tea, error)
	SubscribeOnUpdate(ctx context.Context) (<-chan *model.Tea, error)
	SubscribeOnDelete(ctx context.Context) (<-chan gqlCommon.ID, error)
	Start()
}

type storage interface {
	WriteRecord(ctx context.Context, rec *common.TeaData) (record *common.Tea, err error)
	ReadRecord(ctx context.Context, id uuid.UUID) (record *common.Tea, err error)
	ReadAllRecords(ctx context.Context, search string) ([]common.Tea, error)
	Update(ctx context.Context, id uuid.UUID, rec *common.TeaData) (record *common.Tea, err error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type manager struct {
	storage
	createSubscribers subscribers.TeaSubscribers
	updateSubscribers subscribers.TeaSubscribers
	deleteSubscribers subscribers.IDSubscribers
	create            chan *common.Tea
	update            chan *common.Tea
	delete            chan uuid.UUID
}

func (m *manager) SubscribeOnCreate(ctx context.Context) (<-chan *model.Tea, error) {
	ch := make(chan *model.Tea)
	m.createSubscribers.Push(ctx, ch)
	return ch, nil
}

func (m *manager) SubscribeOnUpdate(ctx context.Context) (<-chan *model.Tea, error) {
	ch := make(chan *model.Tea)
	m.updateSubscribers.Push(ctx, ch)
	return ch, nil
}

func (m *manager) SubscribeOnDelete(ctx context.Context) (<-chan gqlCommon.ID, error) {
	ch := make(chan gqlCommon.ID)
	m.deleteSubscribers.Push(ctx, ch)
	return ch, nil
}

func (m *manager) Get(ctx context.Context, id uuid.UUID) (record *common.Tea, err error) {
	return m.ReadRecord(ctx, id)
}

func (m *manager) List(ctx context.Context, search *string) ([]common.Tea, error) {
	if search == nil {
		return m.ReadAllRecords(ctx, "")
	}
	return m.ReadAllRecords(ctx, *search)
}

func (m *manager) Create(ctx context.Context, data *common.TeaData) (*common.Tea, error) {
	res, err := m.storage.WriteRecord(ctx, data)
	if err != nil {
		return nil, err
	}
	m.create <- res
	return res, nil
}

func (m *manager) Update(ctx context.Context, id uuid.UUID, rec *common.TeaData) (*common.Tea, error) {
	res, err := m.storage.Update(ctx, id, rec)
	if err != nil {
		return nil, err
	}
	m.update <- res
	return res, nil
}

func (m *manager) Delete(ctx context.Context, id uuid.UUID) error {
	if err := m.storage.Delete(ctx, id); err != nil {
		return err
	}
	m.delete <- id
	return nil
}

func (m *manager) Start() {
	go m.loop()
}

func NewManager(storage storage) TeaManager {
	return &manager{
		storage:           storage,
		createSubscribers: subscribers.NewTeaSubscribers(),
		updateSubscribers: subscribers.NewTeaSubscribers(),
		deleteSubscribers: subscribers.NewIDSubscribers(),
		create:            make(chan *common.Tea, 100),
		update:            make(chan *common.Tea, 100),
		delete:            make(chan uuid.UUID, 100),
	}
}
