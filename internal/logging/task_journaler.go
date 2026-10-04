package logging

import (
	"fmt"
	"time"
)

type BatchHandler func([]LogEntry)

type LogEntry struct {
	TaskUUID  string
	Timestamp time.Time
	Source    string
	Message   string
}

type TaskJournalParameters struct {
	MaxRecords    int
	FlushInterval time.Duration
	BatchFn       BatchHandler
}

func (p TaskJournalParameters) Validate() error {
	if p.MaxRecords < 1 {
		return fmt.Errorf("max_records must be >= 1")
	}
	if p.BatchFn == nil {
		return fmt.Errorf("batch handler cannot be nil")
	}
	return nil
}

type TaskJournalMetadata struct {
	TaskUUID string
	Source   string
}

type TaskJournaler struct {
	doneCh        chan struct{}
	inputCh       chan LogEntry
	batchFn       BatchHandler
	maxRecords    int
	flushInterval time.Duration
}

func NewTaskJournaler(params TaskJournalParameters) (*TaskJournaler, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	tj := &TaskJournaler{
		doneCh:        make(chan struct{}),
		inputCh:       make(chan LogEntry, 1<<10),
		maxRecords:    params.MaxRecords,
		flushInterval: params.FlushInterval,
		batchFn:       params.BatchFn,
	}
	go tj.process()
	return tj, nil
}

func (t *TaskJournaler) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Source:    "stream",
		Message:   string(p),
	}
	select {
	case t.inputCh <- entry:
		return len(p), nil
	default:
		return len(p), nil // Drop if full
	}
}

func (t *TaskJournaler) Log(msg string, metadata TaskJournalMetadata) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   msg,
		TaskUUID:  metadata.TaskUUID,
		Source:    metadata.Source,
	}
	select {
	case t.inputCh <- entry:
	default:
	}
}

func (t *TaskJournaler) process() {
	defer close(t.doneCh)

	ticker := time.NewTicker(t.flushInterval)
	defer ticker.Stop()

	buffer := make([]LogEntry, 0, t.maxRecords)

	for {
		select {
		case entry, ok := <-t.inputCh:
			if !ok {
				// sweep remaining logs
				if len(buffer) > 0 {
					t.batchFn(buffer)
				}
				return
			}
			buffer = append(buffer, entry)
			if len(buffer) >= t.maxRecords {
				t.batchFn(buffer)
				buffer = buffer[:0] // reset slice length, keep capacity
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				t.batchFn(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (t *TaskJournaler) Close() error {
	close(t.inputCh)
	<-t.doneCh
	return nil
}

// TaggerTaskJournaler wraps a TaskJournal and injects a specific source tag.
type TaggerTaskJournaler struct {
	logger   *TaskJournaler
	metadata TaskJournalMetadata
}

func NewTaggerTaskJournaler(logger *TaskJournaler, metadata TaskJournalMetadata) *TaggerTaskJournaler {
	return &TaggerTaskJournaler{
		logger:   logger,
		metadata: metadata,
	}
}

// Write satisfies io.Writer. It delegates to the logger with the pre-defined source.
func (tw *TaggerTaskJournaler) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	tw.logger.Log(string(p), tw.metadata)
	return len(p), nil
}
