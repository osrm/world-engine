package cardinal

import (
	"errors"
	"github.com/rotisserie/eris"
	"pkg.world.dev/world-engine/cardinal/filter"
	"pkg.world.dev/world-engine/cardinal/persona"
	"pkg.world.dev/world-engine/cardinal/persona/component"
	"pkg.world.dev/world-engine/cardinal/types"
)

// GetSignerForPersonaTag returns the signer address that has been registered for the given persona tag after the
// given tick. If the engine's tick is less than or equal to the given tick, ErrorCreatePersonaTXsNotProcessed is
// returned. If the given personaTag has no signer address, ErrPersonaTagHasNoSigner is returned.
func (w *World) GetSignerForPersonaTag(personaTag string, tick uint64) (addr string, err error) {
	if tick >= w.CurrentTick() {
		return "", persona.ErrCreatePersonaTxsNotProcessed
	}
	var errs []error
	wCtx := NewReadOnlyWorldContext(w)
	s := NewSearch(wCtx, filter.Exact(component.SignerComponent{}))
	err = s.Each(
		func(id types.EntityID) bool {
			sc, err := GetComponent[component.SignerComponent](wCtx, id)
			if err != nil {
				errs = append(errs, err)
			}
			if sc != nil && sc.PersonaTag == personaTag {
				addr = sc.SignerAddress
				return false
			}
			return true
		},
	)
	errs = append(errs, err)
	if addr == "" {
		return "", persona.ErrPersonaTagHasNoSigner
	}
	return addr, errors.Join(errs...)
}

func (w *World) GetPersonaForEVMAddress(addr string) (string, error) {
	var sc *component.SignerComponent
	wCtx := NewReadOnlyWorldContext(w)
	q := NewSearch(wCtx, filter.Exact(component.SignerComponent{}))
	var getComponentErr error
	searchIterationErr := eris.Wrap(
		q.Each(
			func(id types.EntityID) bool {
				var signerComp *component.SignerComponent
				signerComp, getComponentErr = GetComponent[component.SignerComponent](wCtx, id)
				getComponentErr = eris.Wrap(getComponentErr, "")
				if getComponentErr != nil {
					return false
				}
				for _, authAddr := range signerComp.AuthorizedAddresses {
					if authAddr == addr {
						sc = signerComp
						return false
					}
				}
				return true
			},
		), "",
	)
	if getComponentErr != nil {
		return "", getComponentErr
	}
	if searchIterationErr != nil {
		return "", searchIterationErr
	}
	if sc == nil {
		return "", eris.Errorf("address %s does not have a linked persona tag", addr)
	}
	return sc.PersonaTag, nil
}