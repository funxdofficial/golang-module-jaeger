package domain

// InteractionName merepresentasikan nama interaksi (misal: "Controller InquiryAccount").
type InteractionName struct {
	Value string
}

// NewInteractionName membuat InteractionName dari string.
func NewInteractionName(value string) InteractionName {
	return InteractionName{Value: value}
}

// InteractionTypeName merepresentasikan tipe interaksi (misal: "controllers").
type InteractionTypeName struct {
	Value string
}

// NewInteractionTypeName membuat InteractionTypeName dari string.
func NewInteractionTypeName(value string) InteractionTypeName {
	return InteractionTypeName{Value: value}
}

// NewInteractionTypeType alias untuk NewInteractionTypeName (sesuai naming di requirement).
func NewInteractionTypeType(value string) InteractionTypeName {
	return NewInteractionTypeName(value)
}
