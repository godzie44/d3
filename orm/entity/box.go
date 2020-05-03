package entity

type Box struct {
	Entity interface{}
	Meta   *MetaInfo
}

func NewBox(entity interface{}, meta *MetaInfo) *Box {
	return &Box{Entity: entity, Meta: meta}
}

func (b *Box) ExtractPk() (interface{}, error) {
	return b.Meta.ExtractPkValue(b.Entity)
}

func (b *Box) GetEName() Name {
	return b.Meta.EntityName
}

func (b *Box) GetRelatedMeta(relEntityName Name) *MetaInfo {
	return b.Meta.RelatedMeta[relEntityName]
}
