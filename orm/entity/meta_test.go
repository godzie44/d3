package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateMeta(t *testing.T) {
	meta, _ := CreateMeta((*testEntity)(nil))

	assert.Equal(t, meta.EntityName, Name("d3/orm/entity/testEntity"))
}

func TestCreateMetaFromVariousReflectionsOfOneEntity(t *testing.T) {
	meta1, _ := CreateMeta((*testEntity)(nil))
	meta2, _ := CreateMeta(&testEntity{})

	//cause (*testEntity)(nil) != &testEntity{} do this
	meta2.Tpl = meta1.Tpl
	assert.Equal(t, meta1, meta2)

	meta3, _ := CreateMeta(testEntity{})
	meta3.Tpl = meta1.Tpl

	assert.Equal(t, meta1, meta3)
}
