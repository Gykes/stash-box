package edit

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gofrs/uuid"

	"github.com/stashapp/stash-box/pkg/models"
	"github.com/stashapp/stash-box/pkg/utils"
)

type TagEditProcessor struct {
	mutator
}

func Tag(fac models.Repo, edit *models.Edit) *TagEditProcessor {
	return &TagEditProcessor{
		mutator{
			fac:  fac,
			edit: edit,
		},
	}
}

func (m *TagEditProcessor) Edit(input models.TagEditInput, inputArgs utils.ArgumentsQuery) error {
	var err error
	switch input.Edit.Operation {
	case models.OperationEnumModify:
		err = m.modifyEdit(input, inputArgs)
	case models.OperationEnumMerge:
		err = m.mergeEdit(input, inputArgs)
	case models.OperationEnumDestroy:
		err = m.destroyEdit(input)
	case models.OperationEnumCreate:
		err = m.createEdit(input, inputArgs)
	}

	return err
}

func (m *TagEditProcessor) modifyEdit(input models.TagEditInput, inputArgs utils.ArgumentsQuery) error {
	tqb := m.fac.Tag()

	// get the existing tag
	tagID := *input.Edit.ID
	tag, err := tqb.Find(tagID)

	if err != nil {
		return err
	}

	var entity editEntity = tag
	if err := validateEditEntity(&entity, tagID, "tag"); err != nil {
		return err
	}

	// perform a diff against the input and the current object
	detailArgs := inputArgs.Field("details")
	tagEdit := input.Details.TagEditFromDiff(*tag, detailArgs)

	aliases, err := tqb.GetAliases(tagID)
	if err != nil {
		return err
	}

	if input.Details.Aliases != nil || inputArgs.Field("aliases").IsNull() {
		tagEdit.New.AddedAliases, tagEdit.New.RemovedAliases = utils.SliceCompare(input.Details.Aliases, aliases)
	}

	if reflect.DeepEqual(tagEdit.Old, tagEdit.New) {
		return ErrNoChanges
	}

	return m.edit.SetData(tagEdit)
}

func (m *TagEditProcessor) mergeEdit(input models.TagEditInput, inputArgs utils.ArgumentsQuery) error {
	tqb := m.fac.Tag()

	// get the existing tag
	if input.Edit.ID == nil {
		return ErrMergeIDMissing
	}
	tagID := *input.Edit.ID
	tag, err := tqb.Find(*input.Edit.ID)

	if err != nil {
		return err
	}

	if tag == nil {
		return fmt.Errorf("%w: target tag %s", ErrEntityNotFound, tagID.String())
	}

	var mergeSources []uuid.UUID
	for _, sourceID := range input.Edit.MergeSourceIds {
		sourceTag, err := tqb.Find(sourceID)
		if err != nil {
			return err
		}

		if sourceTag == nil {
			return fmt.Errorf("%w: source tag %s", ErrEntityNotFound, sourceID.String())
		}
		if tagID == sourceID {
			return ErrMergeTargetIsSource
		}
		mergeSources = append(mergeSources, sourceID)
	}

	if len(mergeSources) < 1 {
		return ErrNoMergeSources
	}

	// perform a diff against the input and the current object
	tagEdit := input.Details.TagEditFromMerge(*tag, mergeSources, inputArgs)

	aliases, err := tqb.GetAliases(tagID)

	if err != nil {
		return err
	}

	tagEdit.New.AddedAliases, tagEdit.New.RemovedAliases = utils.SliceCompare(input.Details.Aliases, aliases)

	return m.edit.SetData(tagEdit)
}

func (m *TagEditProcessor) createEdit(input models.TagEditInput, inputArgs utils.ArgumentsQuery) error {
	tagEdit := input.Details.TagEditFromCreate(inputArgs)

	tagEdit.New.AddedAliases = input.Details.Aliases

	return m.edit.SetData(tagEdit)
}

func (m *TagEditProcessor) destroyEdit(input models.TagEditInput) error {
	tqb := m.fac.Tag()

	// Get the existing tag
	tagID := *input.Edit.ID
	tag, err := tqb.Find(tagID)

	if err != nil {
		return err
	}

	var entity editEntity = tag
	return validateEditEntity(&entity, tagID, "tag")
}

func (m *TagEditProcessor) CreateJoin(input models.TagEditInput) error {
	if input.Edit.ID != nil {
		editTag := models.EditTag{
			EditID: m.edit.ID,
			TagID:  *input.Edit.ID,
		}

		return m.fac.Edit().CreateEditTag(editTag)
	}

	return nil
}

func (m *TagEditProcessor) apply() error {
	tqb := m.fac.Tag()
	eqb := m.fac.Edit()
	operation := m.operation()
	isCreate := operation == models.OperationEnumCreate

	var tag *models.Tag
	if !isCreate {
		tagID, err := eqb.FindTagID(m.edit.ID)
		if err != nil {
			return err
		}
		tag, err = tqb.Find(*tagID)
		if err != nil {
			return err
		}
		if tag == nil {
			return fmt.Errorf("%w: tag %s", ErrEntityNotFound, tagID.String())
		}
		tag.UpdatedAt = time.Now()
	}

	newTag, err := tqb.ApplyEdit(*m.edit, operation, tag)
	if err != nil {
		return err
	}

	if isCreate {
		editTag := models.EditTag{
			EditID: m.edit.ID,
			TagID:  newTag.ID,
		}

		err = eqb.CreateEditTag(editTag)
		if err != nil {
			return err
		}
	}

	return nil
}
