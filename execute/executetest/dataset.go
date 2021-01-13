package executetest

import (
	"testing"

	uuid "github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

func RandomDatasetID() execute.DatasetID {
	// uuid.NewV4 can return an error because of enthropy. We will stick with the previous
	// behavior of panicing on errors when creating new uuid's
	return execute.DatasetID(uuid.Must(uuid.NewV4()))
}

type Dataset struct {
	DatasetID             execute.DatasetID
	Retractions           []flux.GroupKey
	ProcessingTimeUpdates []execute.Time
	WatermarkUpdates      []execute.Time
	Finished              bool
	FinishedErr           error
}

func NewDataset(id execute.DatasetID) *Dataset {
	return &Dataset{
		DatasetID: id,
	}
}

func (d *Dataset) ID() execute.DatasetID {
	return d.DatasetID
}

func (d *Dataset) AddTransformation(t execute.Transformation) {
	panic("not implemented")
}

func (d *Dataset) RetractTable(key flux.GroupKey) error {
	d.Retractions = append(d.Retractions, key)
	return nil
}

func (d *Dataset) UpdateProcessingTime(t execute.Time) error {
	d.ProcessingTimeUpdates = append(d.ProcessingTimeUpdates, t)
	return nil
}

func (d *Dataset) UpdateWatermark(mark execute.Time) error {
	d.WatermarkUpdates = append(d.WatermarkUpdates, mark)
	return nil
}

func (d *Dataset) Finish(err error) {
	if d.Finished {
		panic("finish has already been called")
	}
	d.Finished = true
	d.FinishedErr = err
}

func (d *Dataset) SetTriggerSpec(t plan.TriggerSpec) {
	panic("not implemented")
}

type NewTransformation func(execute.Dataset, execute.TableBuilderCache) execute.Transformation

func TransformationPassThroughTestHelper(t *testing.T, newTr NewTransformation) {
	t.Helper()

	now := execute.Now()
	d := NewDataset(RandomDatasetID())
	c := execute.NewTableBuilderCache(UnlimitedAllocator)
	c.SetTriggerSpec(plan.DefaultTriggerSpec)

	parentID := RandomDatasetID()
	tr := newTr(d, c)
	if err := tr.UpdateWatermark(parentID, now); err != nil {
		t.Fatal(err)
	}
	if err := tr.UpdateProcessingTime(parentID, now); err != nil {
		t.Fatal(err)
	}
	tr.Finish(parentID, nil)

	exp := &Dataset{
		DatasetID:             d.DatasetID,
		ProcessingTimeUpdates: []execute.Time{now},
		WatermarkUpdates:      []execute.Time{now},
		Finished:              true,
		FinishedErr:           nil,
	}
	if !cmp.Equal(d, exp) {
		t.Errorf("unexpected dataset -want/+got\n%s", cmp.Diff(exp, d))
	}
}
