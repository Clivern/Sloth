// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schedulerChecks struct {
	db.HealthCheckRepository
	due []*db.HealthCheck
}

func (f *schedulerChecks) ClaimDue(time.Time, int, time.Duration) ([]*db.HealthCheck, error) {
	return f.due, nil
}

type schedulerInstances struct {
	db.AgentInstanceRepository
	instance *db.AgentInstance
}

func (f *schedulerInstances) GetById(db.Id) (*db.AgentInstance, error) {
	return f.instance, nil
}

type schedulerPublisher struct {
	subjects []string
	payloads [][]byte
}

func (f *schedulerPublisher) Publish(subject string, data []byte) error {
	f.subjects = append(f.subjects, subject)
	f.payloads = append(f.payloads, data)
	return nil
}

func (f *schedulerPublisher) Flush() error {
	return nil
}

func TestUnitScheduler(t *testing.T) {
	definition := `{"path":"/healthz","scheme":"http"}`
	checks := &schedulerChecks{due: []*db.HealthCheck{
		{Id: "check-1", InstanceId: "inst-1", Type: db.HealthCheckTypeHTTP, Definition: &definition},
	}}
	instances := &schedulerInstances{instance: &db.AgentInstance{Address: "10.0.12.41", Port: 8080}}
	publisher := &schedulerPublisher{}

	err := NewScheduler(checks, instances, publisher).Tick(time.Now())
	require.NoError(t, err)
	require.Len(t, publisher.payloads, 1)
	assert.Equal(t, []string{HealthCheckRunSubject}, publisher.subjects)

	var job HealthCheckJob
	require.NoError(t, json.Unmarshal(publisher.payloads[0], &job))
	assert.Equal(t, "check-1", job.CheckId)
	assert.Equal(t, "10.0.12.41", job.Address)
	assert.Equal(t, 8080, job.Port)
	assert.JSONEq(t, definition, string(job.Definition))
}
