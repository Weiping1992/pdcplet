//go:build unit

package vmiproxy

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// --- isVmiReady tests ---

func TestIsVmiReady(t *testing.T) {
	tests := []struct {
		name     string
		conds    []kubevirtv1.VirtualMachineInstanceCondition
		expected bool
	}{
		{
			name: "Ready True",
			conds: []kubevirtv1.VirtualMachineInstanceCondition{
				{
					Type:   kubevirtv1.VirtualMachineInstanceReady,
					Status: k8sv1.ConditionTrue,
				},
			},
			expected: true,
		},
		{
			name: "Ready False",
			conds: []kubevirtv1.VirtualMachineInstanceCondition{
				{
					Type:   kubevirtv1.VirtualMachineInstanceReady,
					Status: k8sv1.ConditionFalse,
				},
			},
			expected: false,
		},
		{
			name:     "No Ready Condition",
			conds:    []kubevirtv1.VirtualMachineInstanceCondition{},
			expected: false,
		},
		{
			name: "Other Condition",
			conds: []kubevirtv1.VirtualMachineInstanceCondition{
				{
					Type:   "Other",
					Status: k8sv1.ConditionTrue,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		vmi := &kubevirtv1.VirtualMachineInstance{
			Status: kubevirtv1.VirtualMachineInstanceStatus{
				Conditions: tt.conds,
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isVmiReady(vmi))
		})
	}
}

// --- getNodeName tests ---

func TestGetNodeName(t *testing.T) {
	t.Run("NODE_NAME set", func(t *testing.T) {
		t.Setenv("NODE_NAME", "testnode")
		assert.Equal(t, "testnode", getNodeName())
	})
	t.Run("NODE_NAME not set", func(t *testing.T) {
		t.Setenv("NODE_NAME", "")
		assert.Equal(t, "", getNodeName())
	})
}

// --- doJob tests ---

type mockCache struct {
	setTaskIdCalled   bool
	getTaskIdCalled   bool
	deleteDoneCalled  bool
	taskId            string
	getTaskIdErr      error
}

func (m *mockCache) SetTaskId(name, taskId string) {
	m.setTaskIdCalled = true
	m.taskId = taskId
}
func (m *mockCache) GetTaskId(name string) (string, error) {
	m.getTaskIdCalled = true
	return m.taskId, m.getTaskIdErr
}
func (m *mockCache) DeleteDone(name string) {
	m.deleteDoneCalled = true
}
func (m *mockCache) Update(name string, status interface{}) (bool, bool) { return false, false }
func (m *mockCache) MarkDelete(name string)                              {}

type mockInpClient struct {
	createTaskCalled bool
	closeTaskCalled  bool
	createTaskErr    error
	closeTaskErr     error
	createTaskId     string
}

func (m *mockInpClient) CreateTask(meta map[string]string) (string, error) {
	m.createTaskCalled = true
	return m.createTaskId, m.createTaskErr
}
func (m *mockInpClient) CloseTask(taskId string) error {
	m.closeTaskCalled = true
	return m.closeTaskErr
}

func TestDoJob_CreateTaskOp(t *testing.T) {
	cache := &mockCache{}
	inp := &mockInpClient{createTaskId: "tid123"}
	mod := &vmiProxyModule{
		cache:     cache,
		inpclient: inp,
	}
	vmi := &kubevirtv1.VirtualMachineInstance{}
	item := workqueueItem{vmi: vmi, op: CreateTaskOp}

	mod.doJob(item)
	assert.True(t, inp.createTaskCalled)
	assert.True(t, cache.setTaskIdCalled)
}

func TestDoJob_CreateTaskOp_Error(t *testing.T) {
	cache := &mockCache{}
	inp := &mockInpClient{createTaskErr: errors.New("fail")}
	mod := &vmiProxyModule{
		cache:     cache,
		inpclient: inp,
	}
	vmi := &kubevirtv1.VirtualMachineInstance{}
	item := workqueueItem{vmi: vmi, op: CreateTaskOp}

	mod.doJob(item)
	assert.True(t, inp.createTaskCalled)
	assert.False(t, cache.setTaskIdCalled)
}

func TestDoJob_CloseTaskOp(t *testing.T) {
	cache := &mockCache{taskId: "tid456"}
	inp := &mockInpClient{}
	mod := &vmiProxyModule{
		cache:     cache,
		inpclient: inp,
	}
	vmi := &kubevirtv1.VirtualMachineInstance{}
	item := workqueueItem{vmi: vmi, op: CloseTaskOp}

	mod.doJob(item)
	assert.True(t, cache.getTaskIdCalled)
	assert.True(t, inp.closeTaskCalled)
	assert.True(t, cache.deleteDoneCalled)
}

func TestDoJob_CloseTaskOp_GetTaskIdError(t *testing.T) {
	cache := &mockCache{getTaskIdErr: errors.New("not found")}
	inp := &mockInpClient{}
	mod := &vmiProxyModule{
		cache:     cache,
		inpclient: inp,
	}
	vmi := &kubevirtv1.VirtualMachineInstance{}
	item := workqueueItem{vmi: vmi, op: CloseTaskOp}

	mod.doJob(item)
	assert.True(t, cache.getTaskIdCalled)
	assert.False(t, inp.closeTaskCalled)
	assert.True(t, cache.deleteDoneCalled)
}

func TestDoJob_CloseTaskOp_CloseTaskError(t *testing.T) {
	cache := &mockCache{taskId: "tid789"}
	inp := &mockInpClient{closeTaskErr: errors.New("fail")}
	mod := &vmiProxyModule{
		cache:     cache,
		inpclient: inp,
	}
	vmi := &kubevirtv1.VirtualMachineInstance{}
	item := workqueueItem{vmi: vmi, op: CloseTaskOp}

	mod.doJob(item)
	assert.True(t, cache.getTaskIdCalled)
	assert.True(t, inp.closeTaskCalled)
	assert.True(t, cache.deleteDoneCalled)
}