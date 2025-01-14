package counterparty

import (
	"github.com/iota-uz/iota-sdk/modules/core/domain/value_objects/tax"
	"time"
)

type Counterparty interface {
	ID() uint
	SetID(uint)

	Tin() tax.Tin
	SetTin(t tax.Tin)

	Name() string
	SetName(string)

	Type() Type
	SetType(Type)

	LegalType() LegalType
	SetLegalType(LegalType)

	LegalAddress() string
	SetLegalAddress(string)

	CreatedAt() time.Time
	UpdatedAt() time.Time
}
