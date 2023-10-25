package cardinal_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"pkg.world.dev/world-engine/cardinal"
	"pkg.world.dev/world-engine/cardinal/test_utils"
)

type AddHealthToEntityTx struct {
	TargetID cardinal.EntityID
	Amount   int
}

type AddHealthToEntityResult struct{}

var addHealthToEntity = cardinal.NewTransactionType[AddHealthToEntityTx, AddHealthToEntityResult]("add_health")

func TestTransactionExample(t *testing.T) {
	world, doTick := test_utils.MakeWorldAndTicker(t)
	assert.NilError(t, cardinal.RegisterComponent[Health](world))
	assert.NilError(t, cardinal.RegisterTransactions(world, addHealthToEntity))
	cardinal.RegisterSystems(world, func(worldCtx cardinal.WorldContext) error {
		for _, tx := range addHealthToEntity.In(worldCtx) {
			targetID := tx.Value().TargetID
			err := cardinal.UpdateComponent[Health](worldCtx, targetID, func(h *Health) *Health {
				h.Value = tx.Value().Amount
				return h
			})
			assert.Check(t, err == nil)
		}
		return nil
	})

	testWorldCtx := test_utils.WorldToWorldContext(world)
	ids, err := cardinal.CreateMany(testWorldCtx, 10, Health{})
	assert.NilError(t, err)

	// Queue up the transaction.
	idToModify := ids[3]
	amountToModify := 20

	test_utils.AddTransactionToWorldByAnyTransaction(world, addHealthToEntity, AddHealthToEntityTx{idToModify, amountToModify})

	// The health change should be applied during this tick
	doTick()

	// Make sure the target entity had its health updated.
	for _, id := range ids {
		health, err := cardinal.GetComponent[Health](testWorldCtx, id)
		assert.NilError(t, err)
		if id == idToModify {
			assert.Equal(t, amountToModify, health.Value)
		} else {
			assert.Equal(t, 0, health.Value)
		}
	}
}
