package simsdk

import (
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

func ToProtoManifest(m Manifest) *simsdkrpc.Manifest {
	proto := &simsdkrpc.Manifest{
		Name:    m.Name,
		Version: m.Version,
	}

	for _, mt := range m.MessageTypes {
		proto.MessageTypes = append(proto.MessageTypes, toProtoMessageType(mt))
	}
	for _, ct := range m.ControlFunctions {
		proto.ControlFunctions = append(proto.ControlFunctions, toProtoControlFunction(ct))
	}
	for _, cmp := range m.ComponentTypes {
		proto.ComponentTypes = append(proto.ComponentTypes, &simsdkrpc.ComponentType{
			Id:                        cmp.ID,
			DisplayName:               cmp.DisplayName,
			Internal:                  cmp.Internal,
			Description:               cmp.Description,
			SupportsMultipleInstances: cmp.SupportsMultipleInstances, // default to true; adjust as needed
		})
	}
	proto.TransportTypes = toProtoTransportTypes(m.TransportTypes)

	return proto
}

func FromProtoManifest(p *simsdkrpc.Manifest) Manifest {
	return Manifest{
		Name:             p.Name,
		Version:          p.Version,
		MessageTypes:     fromProtoMessageTypes(p.MessageTypes),
		ControlFunctions: fromProtoControlFunctions(p.ControlFunctions),
		ComponentTypes:   fromProtoComponentTypes(p.ComponentTypes),
		TransportTypes:   fromProtoTransportTypes(p.TransportTypes),
	}
}

func toProtoMessageType(mt MessageType) *simsdkrpc.MessageType {
	return &simsdkrpc.MessageType{
		Id:          mt.ID,
		DisplayName: mt.DisplayName,
		Description: mt.Description,
		Fields:      toProtoFieldSpecs(mt.Fields),
	}
}

func fromProtoMessageTypes(proto []*simsdkrpc.MessageType) []MessageType {
	var result []MessageType
	for _, mt := range proto {
		result = append(result, MessageType{
			ID:          mt.Id,
			DisplayName: mt.DisplayName,
			Description: mt.Description,
			Fields:      fromProtoFieldSpecs(mt.Fields),
		})
	}
	return result
}

func toProtoControlFunction(cf ControlFunctionType) *simsdkrpc.ControlFunctionType {
	return &simsdkrpc.ControlFunctionType{
		Id:          cf.ID,
		DisplayName: cf.DisplayName,
		Description: cf.Description,
		Fields:      toProtoFieldSpecs(cf.Fields),
	}
}

func fromProtoControlFunctions(proto []*simsdkrpc.ControlFunctionType) []ControlFunctionType {
	var result []ControlFunctionType
	for _, cf := range proto {
		result = append(result, ControlFunctionType{
			ID:          cf.Id,
			DisplayName: cf.DisplayName,
			Description: cf.Description,
			Fields:      fromProtoFieldSpecs(cf.Fields),
		})
	}
	return result
}

func toProtoTransportTypes(tt []TransportType) []*simsdkrpc.TransportType {
	var result []*simsdkrpc.TransportType
	for _, t := range tt {
		result = append(result, &simsdkrpc.TransportType{
			Id:          t.ID,
			DisplayName: t.DisplayName,
			Description: t.Description,
			Internal:    t.Internal,
		})
	}
	return result
}

func fromProtoTransportTypes(proto []*simsdkrpc.TransportType) []TransportType {
	var result []TransportType
	for _, t := range proto {
		result = append(result, TransportType{
			ID:          t.Id,
			DisplayName: t.DisplayName,
			Description: t.Description,
			Internal:    t.Internal,
		})
	}
	return result
}

func toProtoFieldSpecs(fields []FieldSpec) []*simsdkrpc.FieldSpec {
	var out []*simsdkrpc.FieldSpec
	for _, f := range fields {
		out = append(out, toProtoFieldSpec(f))
	}
	return out
}

func toProtoFieldSpec(f FieldSpec) *simsdkrpc.FieldSpec {
	field := &simsdkrpc.FieldSpec{
		Name:         f.Name,
		Type:         toProtoFieldType(f.Type),
		Required:     f.Required,
		EnumValues:   f.EnumValues,
		Repeated:     f.Repeated,
		Description:  f.Description,
		ObjectFields: toProtoFieldSpecs(f.ObjectFields),
	}
	if f.Subtype != nil {
		field.Subtype = toProtoFieldType(*f.Subtype)
	}
	return field
}

func fromProtoFieldSpecs(proto []*simsdkrpc.FieldSpec) []FieldSpec {
	var result []FieldSpec
	for _, f := range proto {
		result = append(result, fromProtoFieldSpec(f))
	}
	return result
}

func fromProtoFieldSpec(p *simsdkrpc.FieldSpec) FieldSpec {
	field := FieldSpec{
		Name:         p.Name,
		Type:         fromProtoFieldType(p.Type),
		Required:     p.Required,
		EnumValues:   p.EnumValues,
		Repeated:     p.Repeated,
		Description:  p.Description,
		ObjectFields: fromProtoFieldSpecs(p.ObjectFields),
	}
	if p.Subtype != simsdkrpc.FieldType_FIELD_TYPE_UNSPECIFIED {
		sub := fromProtoFieldType(p.Subtype)
		field.Subtype = &sub
	}
	return field
}

func toProtoFieldType(ft FieldType) simsdkrpc.FieldType {
	switch ft {
	case FieldString:
		return simsdkrpc.FieldType_STRING
	case FieldInt:
		return simsdkrpc.FieldType_INT
	case FieldUint:
		return simsdkrpc.FieldType_UINT
	case FieldFloat:
		return simsdkrpc.FieldType_FLOAT
	case FieldBool:
		return simsdkrpc.FieldType_BOOL
	case FieldEnum:
		return simsdkrpc.FieldType_ENUM
	case FieldTimestamp:
		return simsdkrpc.FieldType_TIMESTAMP
	case FieldRepeated:
		return simsdkrpc.FieldType_REPEATED
	case FieldObject:
		return simsdkrpc.FieldType_OBJECT
	default:
		return simsdkrpc.FieldType_FIELD_TYPE_UNSPECIFIED
	}
}

func fromProtoFieldType(ft simsdkrpc.FieldType) FieldType {
	switch ft {
	case simsdkrpc.FieldType_STRING:
		return FieldString
	case simsdkrpc.FieldType_INT:
		return FieldInt
	case simsdkrpc.FieldType_UINT:
		return FieldUint
	case simsdkrpc.FieldType_FLOAT:
		return FieldFloat
	case simsdkrpc.FieldType_BOOL:
		return FieldBool
	case simsdkrpc.FieldType_ENUM:
		return FieldEnum
	case simsdkrpc.FieldType_TIMESTAMP:
		return FieldTimestamp
	case simsdkrpc.FieldType_REPEATED:
		return FieldRepeated
	case simsdkrpc.FieldType_OBJECT:
		return FieldObject
	default:
		return FieldType("FIELD_TYPE_UNSPECIFIED")
	}
}

func fromProtoComponentTypes(proto []*simsdkrpc.ComponentType) []ComponentType {
	var result []ComponentType
	for _, ct := range proto {
		result = append(result, ComponentType{
			ID:                        ct.Id,
			DisplayName:               ct.DisplayName,
			Internal:                  ct.Internal,
			Description:               ct.Description,
			SupportsMultipleInstances: ct.SupportsMultipleInstances,
		})
	}
	return result
}

func ToProtoSimMessage(m *SimMessage) *simsdkrpc.SimMessage {
	if m == nil {
		return nil
	}
	return &simsdkrpc.SimMessage{
		MessageType: m.MessageType,
		MessageId:   m.MessageID,
		ComponentId: m.ComponentID,
		Payload:     m.Payload,
		Metadata:    m.Metadata,
	}
}

func FromProtoSimMessage(p *simsdkrpc.SimMessage) *SimMessage {
	if p == nil {
		return nil
	}
	return &SimMessage{
		MessageType: p.MessageType,
		MessageID:   p.MessageId,
		ComponentID: p.ComponentId,
		Payload:     p.Payload,
		Metadata:    p.Metadata,
	}
}
