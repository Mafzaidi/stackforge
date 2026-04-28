package serializer

import "github.com/mafzaidi/stackforge/internal/domain/entity"

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func FromTag(e *entity.Tag) Tag {
	return Tag{
		ID:   e.ID,
		Name: e.Name,
	}
}

func FromTagList(entities []*entity.Tag) []Tag {
	result := make([]Tag, len(entities))
	for i, e := range entities {
		result[i] = FromTag(e)
	}
	return result
}
