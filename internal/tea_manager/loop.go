package tea_manager

import model "github.com/lueurxax/teaelephantmemory/internal/server/api/v2/models"

func (m *manager) loop() {
	for {
		select {
		case tea := <-m.create:
			m.muCreate.RLock()
			for _, el := range m.createSubscribers {
				el <- model.FromCommonTea(tea)
			}
			m.muCreate.RUnlock()
		case tea := <-m.update:
			m.muUpdate.RLock()
			for _, el := range m.updateSubscribers {
				el <- model.FromCommonTea(tea)
			}
			m.muUpdate.RUnlock()
		case id := <-m.delete:
			m.muDelete.RLock()
			for _, el := range m.deleteSubscribers {
				el <- id
			}
			m.muDelete.RUnlock()
		}
	}
}