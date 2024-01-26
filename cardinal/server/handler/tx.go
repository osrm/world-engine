package handler

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	"pkg.world.dev/world-engine/cardinal/ecs"
	"pkg.world.dev/world-engine/cardinal/types/message"
	"pkg.world.dev/world-engine/sign"
)

var (
	ErrNoPersonaTag               = errors.New("persona tag is required")
	ErrWrongNamespace             = errors.New("incorrect namespace")
	ErrSystemTransactionRequired  = errors.New("system transaction required")
	ErrSystemTransactionForbidden = errors.New("system transaction forbidden")
)

// PostTransactionResponse is the HTTP response for a successful transaction submission
type PostTransactionResponse struct {
	TxHash string
	Tick   uint64
}

//nolint:gocognit
func PostTransaction(
	msgs map[string]map[string]message.Message, engine *ecs.Engine, disableSigVerification bool,
) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		msgType, ok := msgs[ctx.Params("group")][ctx.Params("name")]
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, "message type not found")
		}

		// Parse the request body into a sign.Transaction struct
		tx := new(sign.Transaction)
		if err := ctx.BodyParser(tx); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "failed to parse request body: "+err.Error())
		}

		// Validate the transaction
		if err := validateTx(tx); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid transaction payload: "+err.Error())
		}

		// Decode the message from the transaction
		msg, err := msgType.Decode(tx.Body)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "failed to decode message from transaction")
		}

		var signerAddress string
		// TODO(scott): this should be refactored; I don't see why getting signer address needs to be different here,
		//  both of them should just derive the signer address using ecrecover from signature
		if msgType.Name() == ecs.CreatePersonaMsg.Name() {
			// don't need to check the cast bc we already validated this above
			createPersonaMsg, _ := msg.(ecs.CreatePersona)
			signerAddress = createPersonaMsg.SignerAddress
		} else {
			signerAddress, err = engine.GetSignerForPersonaTag(tx.PersonaTag, 0)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "could not get signer for persona: "+err.Error())
			}
		}

		// If signature verification is enabled, validate the transaction
		if !disableSigVerification {
			if err = validateSignature(tx, signerAddress, engine.Namespace().String(),
				tx.IsSystemTransaction()); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "failed to validate transaction: "+err.Error())
			}
			// TODO(scott): this should be refactored; it should be the responsibility of the engine tx processor
			//  to mark the nonce as used once it's included in the tick, not the server.
			if err = engine.UseNonce(signerAddress, tx.Nonce); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed to use nonce: "+err.Error())
			}
		}

		// Add the transaction to the engine
		// TODO(scott): this should just deal with txpool instead of having to go through engine
		tick, hash := engine.AddTransaction(msgType.ID(), msg, tx)

		return ctx.JSON(&PostTransactionResponse{
			TxHash: string(hash),
			Tick:   tick,
		})
	}
}

// validateTx validates the transaction payload
func validateTx(tx *sign.Transaction) error {
	// TODO(scott): we should use the validator package here
	if tx.PersonaTag == "" {
		return ErrNoPersonaTag
	}
	return nil
}

// validateSignature validates that the signature of transaction is valid
func validateSignature(tx *sign.Transaction, signerAddr string, namespace string, systemTx bool) error {
	if tx.Namespace != namespace {
		return eris.Wrap(ErrWrongNamespace, fmt.Sprintf("expected %q got %q", namespace, tx.Namespace))
	}
	if systemTx && !tx.IsSystemTransaction() {
		return eris.Wrap(ErrSystemTransactionRequired, "")
	}
	if !systemTx && tx.IsSystemTransaction() {
		return eris.Wrap(ErrSystemTransactionForbidden, "")
	}
	return eris.Wrap(tx.Verify(signerAddr), "")
}