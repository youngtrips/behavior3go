package core

import (
	"fmt"
	"reflect"
)

/**
 * The Blackboard is the memory structure required by `BehaviorTree` and its
 * nodes. It only have 2 public methods: `set` and `get`. These methods works
 * in 3 different contexts: global, per tree, and per node per tree.
 *
 * Suppose you have two different trees controlling a single object with a
 * single blackboard, then:
 *
 * - In the global context, all nodes will access the stored information.
 * - In per tree context, only nodes sharing the same tree share the stored
 *   information.
 * - In per node per tree context, the information stored in the blackboard
 *   can only be accessed by the same node that wrote the data.
 *
 * The context is selected indirectly by the parameters provided to these
 * methods, for example:
 *
 *     // getting/setting variable in global context
 *     blackboard.set('testKey', 'value');
 *     var value = blackboard.get('testKey');
 *
 *     // getting/setting variable in per tree context
 *     blackboard.set('testKey', 'value', tree.id);
 *     var value = blackboard.get('testKey', tree.id);
 *
 *     // getting/setting variable in per node per tree context
 *     blackboard.set('testKey', 'value', tree.id, node.id);
 *     var value = blackboard.get('testKey', tree.id, node.id);
 *
 * Note: Internally, the blackboard store these memories in different
 * objects, being the global on `_baseMemory`, the per tree on `_treeMemory`
 * and the per node per tree dynamically create inside the per tree memory
 * (it is accessed via `_treeMemory[id].nodeMemory`). Avoid to use these
 * variables manually, use `get` and `set` instead.
 *
 * @module b3
 * @class Blackboard
**/
//------------------------TreeData-------------------------
type TreeData struct {
	NodeMemory     *Memory
	OpenNodes      []IBaseNode
	TraversalDepth int
	TraversalCycle int
}

func NewTreeData() *TreeData {
	return &TreeData{NewMemory(), make([]IBaseNode, 0), 0, 0}
}

//------------------------Memory-------------------------
type Memory struct {
	_memory map[string]interface{}
}

func NewMemory() *Memory {
	return &Memory{make(map[string]interface{})}
}

func (this *Memory) Get(key string) interface{} {
	return this._memory[key]
}
func (this *Memory) Set(key string, val interface{}) {
	this._memory[key] = val
}
func (this *Memory) Remove(key string) {
	delete(this._memory, key)
}

//------------------------TreeMemory-------------------------
type TreeMemory struct {
	*Memory
	_treeData   *TreeData
	_nodeMemory map[string]*Memory
}

func NewTreeMemory() *TreeMemory {
	return &TreeMemory{NewMemory(), NewTreeData(), make(map[string]*Memory)}
}

type Storage interface {
	Set(key string, value interface{}, treeScope string, nodeScope string)
	Remove(key string, treeScope string, nodeScope string)
	Foreach(func(key string, value interface{}, treeScope string, nodeScope string))
}

//------------------------Blackboard-------------------------
type Blackboard struct {
	_storage    Storage
	_baseMemory *Memory
	_treeMemory map[string]*TreeMemory
}

func NewBlackboard(storage Storage) *Blackboard {
	p := &Blackboard{
		_storage: storage,
	}
	p.Initialize()
	return p
}

func (this *Blackboard) Initialize() {
	this._baseMemory = NewMemory()
	this._treeMemory = make(map[string]*TreeMemory)
	if this._storage != nil {
		this._storage.Foreach(func(key string, value interface{}, treeScope string, nodeScope string) {
			if treeScope != "" && nodeScope != "" {
				this.Set(key, value, treeScope, nodeScope)
			} else if treeScope != "" {
				this.SetTree(key, value, treeScope)
			} else {
				this.SetMem(key, value)
			}
		})
	}
}

/**
 * Internal method to retrieve the tree context memory. If the memory does
 * not exist, this method creates it.
 *
 * @method _getTreeMemory
 * @param {string} treeScope The id of the tree in scope.
 * @return {Object} The tree memory.
 * @protected
**/
func (this *Blackboard) _getTreeMemory(treeScope string) *TreeMemory {
	if _, ok := this._treeMemory[treeScope]; !ok {
		this._treeMemory[treeScope] = NewTreeMemory()
	}
	return this._treeMemory[treeScope]
}

/**
 * Internal method to retrieve the node context memory, given the tree
 * memory. If the memory does not exist, this method creates is.
 *
 * @method _getNodeMemory
 * @param {String} treeMemory the tree memory.
 * @param {String} nodeScope The id of the node in scope.
 * @return {Object} The node memory.
 * @protected
**/
func (this *Blackboard) _getNodeMemory(treeMemory *TreeMemory, nodeScope string) *Memory {
	memory := treeMemory._nodeMemory
	if _, ok := memory[nodeScope]; !ok {
		memory[nodeScope] = NewMemory()
	}

	return memory[nodeScope]
}

/**
 * Internal method to retrieve the context memory. If treeScope and
 * nodeScope are provided, this method returns the per node per tree
 * memory. If only the treeScope is provided, it returns the per tree
 * memory. If no parameter is provided, it returns the global memory.
 * Notice that, if only nodeScope is provided, this method will still
 * return the global memory.
 *
 * @method _getMemory
 * @param {String} treeScope The id of the tree scope.
 * @param {String} nodeScope The id of the node scope.
 * @return {Object} A memory object.
 * @protected
**/
func (this *Blackboard) _getMemory(treeScope, nodeScope string) *Memory {
	var memory = this._baseMemory

	if len(treeScope) > 0 {
		treeMem := this._getTreeMemory(treeScope)
		memory = treeMem.Memory
		if len(nodeScope) > 0 {
			memory = this._getNodeMemory(treeMem, nodeScope)
		}
	}

	return memory
}

/**
 * Stores a value in the blackboard. If treeScope and nodeScope are
 * provided, this method will save the value into the per node per tree
 * memory. If only the treeScope is provided, it will save the value into
 * the per tree memory. If no parameter is provided, this method will save
 * the value into the global memory. Notice that, if only nodeScope is
 * provided (but treeScope not), this method will still save the value into
 * the global memory.
 *
 * @method set
 * @param {String} key The key to be stored.
 * @param {String} value The value to be stored.
 * @param {String} treeScope The tree id if accessing the tree or node
 *                           memory.
 * @param {String} nodeScope The node id if accessing the node memory.
**/
func (this *Blackboard) Set(key string, value interface{}, treeScope, nodeScope string) {
	var memory = this._getMemory(treeScope, nodeScope)
	memory.Set(key, value)
	if this._storage != nil {
		this._storage.Set(key, value, treeScope, nodeScope)
	}
}

func (this *Blackboard) SetMem(key string, value interface{}) {
	var memory = this._getMemory("", "")
	memory.Set(key, value)
	if this._storage != nil {
		this._storage.Set(key, value, "", "")
	}
}

func (this *Blackboard) Remove(key string) {
	var memory = this._getMemory("", "")
	memory.Remove(key)
	if this._storage != nil {
		this._storage.Remove(key, "", "")
	}
}
func (this *Blackboard) SetTree(key string, value interface{}, treeScope string) {
	var memory = this._getMemory(treeScope, "")
	memory.Set(key, value)

	if this._storage != nil {
		this._storage.Set(key, value, treeScope, "")
	}
}

func (this *Blackboard) _getTreeData(treeScope string) *TreeData {
	treeMem := this._getTreeMemory(treeScope)
	return treeMem._treeData
}

/**
 * Retrieves a value in the blackboard. If treeScope and nodeScope are
 * provided, this method will retrieve the value from the per node per tree
 * memory. If only the treeScope is provided, it will retrieve the value
 * from the per tree memory. If no parameter is provided, this method will
 * retrieve from the global memory. If only nodeScope is provided (but
 * treeScope not), this method will still try to retrieve from the global
 * memory.
 *
 * @method get
 * @param {String} key The key to be retrieved.
 * @param {String} treeScope The tree id if accessing the tree or node
 *                           memory.
 * @param {String} nodeScope The node id if accessing the node memory.
 * @return {Object} The value stored or undefined.
**/
func (this *Blackboard) Get(key, treeScope, nodeScope string) interface{} {
	memory := this._getMemory(treeScope, nodeScope)
	return memory.Get(key)
}
func (this *Blackboard) GetMem(key string) interface{} {
	memory := this._getMemory("", "")
	return memory.Get(key)
}
func (this *Blackboard) GetFloat64(key, treeScope, nodeScope string) float64 {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(float64)
}
func (this *Blackboard) GetBool(key, treeScope, nodeScope string) bool {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return false
	}
	return v.(bool)
}
func (this *Blackboard) GetInt(key, treeScope, nodeScope string) int {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(int)
}
func (this *Blackboard) GetInt64(key, treeScope, nodeScope string) int64 {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(int64)
}
func (this *Blackboard) GetUInt64(key, treeScope, nodeScope string) uint64 {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(uint64)
}

func (this *Blackboard) GetInt64Safe(key, treeScope, nodeScope string) int64 {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return ReadNumberToInt64(v)
}
func (this *Blackboard) GetUInt64Safe(key, treeScope, nodeScope string) uint64 {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return ReadNumberToUInt64(v)
}

func (this *Blackboard) GetInt32(key, treeScope, nodeScope string) int32 {
	v := this.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(int32)
}

func ReadNumberToInt64(v interface{}) int64 {
	var ret int64
	switch tvalue := v.(type) {
	case uint64:
		ret = int64(tvalue)
	default:
		panic(fmt.Sprintf("错误的类型转成Int64 %v:%+v", reflect.TypeOf(v), v))
	}

	return ret
}

func ReadNumberToUInt64(v interface{}) uint64 {
	var ret uint64
	switch tvalue := v.(type) {
	case int64:
		ret = uint64(tvalue)
	default:
		panic(fmt.Sprintf("错误的类型转成UInt64 %v:%+v", reflect.TypeOf(v), v))
	}
	return ret
}

//
//func ReadNumberToInt32(v interface{}) int32 {
//	var ret int32
//	switch tvalue := v.(type) {
//	case uint16, int16,uint32, int32,uint64,int64,uint16, int16,int:
//		ret = int32(tvalue)
//	default:
//		panic(fmt.Sprintf("错误的类型转成Int32 %v:%+v", reflect.TypeOf(v), v))
//	}
//	return ret
//}
//
//func ReadNumberToUInt32(v interface{}) uint32 {
//	var ret uint32
//	switch tvalue := v.(type) {
//	case uint16, int16,uint32, int32,uint64,int64,uint16, int16,int:
//		ret = uint32(tvalue)
//	default:
//		panic(fmt.Sprintf("错误的类型转成UInt32 %v:%+v", reflect.TypeOf(v), v))
//	}
//	return ret
//}
//
//
//func ReadNumberToInt(v interface{}) int {
//	var ret int
//	switch tvalue := v.(type) {
//	case uint16, int16,uint32, int32,uint64,int64,uint16, int16,int:
//		ret = int(tvalue)
//	default:
//		panic(fmt.Sprintf("int %v:%+v", reflect.TypeOf(v), v))
//	}
//	return ret
//}
