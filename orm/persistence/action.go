package persistence

import (
	"d3/orm/entity"
	"fmt"
)

type actionField struct {
	name string
	val  interface{}
}

func ActionField(name string, val interface{}) *actionField {
	return &actionField{name, val}
}

type (
	databaseAction interface {
		setTableName(name string)
		setFields(fields ...*actionField)
		hasField(needle *actionField) bool
	}

	executableAction interface {
		wasExecuted() bool
		exec(pusher Pusher) error
	}

	waiter interface {
		subscriptionsResolved() bool
		subscriptionsCount() int
		addSubscription(executableAction)
	}

	persistNode interface {
		databaseAction
		executableAction
		waiter
	}

	CompositeAction interface {
		persistNode
		children() []CompositeAction
		addChild(CompositeAction)
		equalTo(CompositeAction) bool
		hasChild(CompositeAction) bool
		//mergeFields like setFields, but set only if field is unique for aggregator and in all childrenActions.
		mergeFields(fields ...*actionField)
		//hasFieldRecursive checks that field exists in the action and in all childrenActions.
		hasFieldRecursive(needle *actionField) bool
	}
)

type baseAction struct {
	TableName string

	Values map[string]interface{}

	executed      bool
	subscriptions []executableAction

	childrenActions []CompositeAction
}

func (b *baseAction) exec(_ Pusher) error {
	b.executed = true
	return nil
}

func (b *baseAction) subscriptionsCount() int {
	return len(b.subscriptions)
}

func (b *baseAction) setTableName(name string) {
	b.TableName = name
}

func (b *baseAction) setFields(fields ...*actionField) {
	for i := range fields {
		b.Values[fields[i].name] = fields[i].val
	}
}

func (b *baseAction) hasField(needle *actionField) bool {
	if val, exists := b.Values[needle.name]; exists {
		equalerVal, isEq := val.(equaler)
		needleEqualer, needleIsEq := needle.val.(equaler)

		if isEq != needleIsEq {
			return false
		}

		if isEq && needleIsEq {
			return equalerVal.equal(needleEqualer)
		}

		return val == needle.val
	}

	return false
}

func (b *baseAction) wasExecuted() bool {
	return b.executed
}

func (b *baseAction) subscriptionsResolved() bool {
	for _, sub := range b.subscriptions {
		if !sub.wasExecuted() {
			return false
		}
	}
	return true
}

func (b *baseAction) addSubscription(act executableAction) {
	b.subscriptions = append(b.subscriptions, act)
}

func (b *baseAction) children() []CompositeAction {
	return b.childrenActions
}

func (b *baseAction) addChild(child CompositeAction) {
	b.childrenActions = append(b.childrenActions, child)
	child.addSubscription(b)
}

func (b *baseAction) hasChild(child CompositeAction) bool {
	for _, c := range b.childrenActions {
		if c == child || c.equalTo(child) {
			return true
		}
	}

	return false
}

func (b *baseAction) mergeFields(fields ...*actionField) {
	for _, field := range fields {
		if !b.hasFieldRecursive(field) {
			b.setFields(field)
		}
	}
}

func (b *baseAction) hasFieldRecursive(field *actionField) bool {
	if b.hasField(field) {
		return true
	}

	for _, child := range b.children() {
		if child.hasFieldRecursive(field) {
			return true
		}
	}

	return false
}

func (b *baseAction) prepareValues() error {
	for name, val := range b.Values {
		if valPromise, ok := val.(*promise); ok {
			v, err := valPromise.unwrap()
			if err != nil {
				return fmt.Errorf("%w", err)
			}
			b.Values[name] = v
		}
	}
	return nil
}

func (b *baseAction) splitValues() ([]string, []interface{}) {
	columns := make([]string, 0, len(b.Values))
	values := make([]interface{}, 0, len(b.Values))

	for name := range b.Values {
		columns = append(columns, name)
		values = append(values, b.Values[name])
	}

	return columns, values
}

type InsertAction struct {
	baseAction
	pkHydrateFn func([]interface{}) error

	pkGenStrategy entity.PkStrategy
	pkCols        []string

	box *persistBox
}

func NewInsertAction(pkHydrator func([]interface{}) error, box *persistBox) *InsertAction {
	action := &InsertAction{pkHydrateFn: pkHydrator, box: box, baseAction: baseAction{Values: make(map[string]interface{})}}

	if box != nil {
		action.pkGenStrategy = box.Meta.Pk.Strategy
		action.pkCols = []string{box.Meta.Pk.Field.DbAlias}
	}

	return action
}

func (i *InsertAction) Box() *entity.Box {
	if i.box == nil {
		return nil
	}
	return i.box.Box
}

func (i *InsertAction) exec(pusher Pusher) error {
	if i.pkGenStrategy == entity.Auto {
		for _, pkCol := range i.pkCols {
			delete(i.Values, pkCol)
		}
	}

	if err := i.prepareValues(); err != nil {
		return fmt.Errorf("insert execution failed: %w", err)
	}

	columns, values := i.splitValues()

	if i.pkHydrateFn != nil && i.pkGenStrategy == entity.Auto {
		tpl := i.box.Meta.CreateKeyTpl()
		err := pusher.InsertWithReturn(i.TableName, columns, values, i.pkCols, func(scanner Scanner) error {
			return scanner.Scan(tpl.Projection()...)
		})
		if err != nil {
			return err
		}

		if err = i.pkHydrateFn(tpl.Key()); err != nil {
			return err
		}
	} else {
		err := pusher.Insert(i.TableName, columns, values)
		if err != nil {
			return err
		}
	}

	return i.baseAction.exec(pusher)
}

func (i *InsertAction) equalTo(agr CompositeAction) bool {
	if action, ok := agr.(*InsertAction); ok {
		if action.pkGenStrategy == entity.Auto {
			return false
		}

		if action.TableName == i.TableName && mapEquals(i.Values, action.Values) {
			return true
		}
	}

	return false
}

type UpdateAction struct {
	baseAction
	identityCondition map[string]interface{}
}

func NewUpdateAction(identityCondition map[string]interface{}) *UpdateAction {
	return &UpdateAction{identityCondition: identityCondition, baseAction: baseAction{Values: make(map[string]interface{})}}
}

func (u *UpdateAction) exec(pusher Pusher) error {
	if len(u.Values) == 0 {
		return u.baseAction.exec(pusher)
	}

	if err := u.prepareValues(); err != nil {
		return fmt.Errorf("update execution failed: %w", err)
	}
	columns, values := u.splitValues()

	for col, val := range u.identityCondition {
		if valPromise, ok := val.(*promise); ok {
			v, err := valPromise.unwrap()
			if err != nil {
				return fmt.Errorf("persist failed: %w", err)
			}
			u.identityCondition[col] = v
		}
	}

	if err := pusher.Update(u.TableName, columns, values, u.identityCondition); err != nil {
		return err
	}

	return u.baseAction.exec(pusher)
}

func (u *UpdateAction) equalTo(act CompositeAction) bool {
	if action, ok := act.(*UpdateAction); ok {
		if action.TableName == u.TableName && mapEquals(u.identityCondition, action.identityCondition) && mapEquals(u.Values, action.Values) {
			return true
		}
	}

	return false
}

type DeleteAction struct {
	baseAction
	deleteCondition map[string]interface{}
}

func NewDeleteAction(deleteCondition map[string]interface{}) *DeleteAction {
	return &DeleteAction{deleteCondition: deleteCondition}
}

func (d *DeleteAction) equalTo(act CompositeAction) bool {
	if action, ok := act.(*DeleteAction); ok {
		if action.TableName == d.TableName && mapEquals(d.deleteCondition, action.deleteCondition) {
			return true
		}
	}

	return false
}

func (d *DeleteAction) exec(pusher Pusher) error {
	err := pusher.Remove(d.TableName, d.deleteCondition)
	if err != nil {
		return err
	}

	return d.baseAction.exec(pusher)
}
