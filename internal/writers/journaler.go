package writers

import (
	"fmt"
	"time"
)

type BatchHandler func([]JournalEntry)

type JournalEntry struct {
	StreamUUID string
	Timestamp  time.Time
	Source     string
	Message    string
}

type JournalerParameters struct {
	MaxRecords    int
	FlushInterval time.Duration
	BatchFn       BatchHandler
}

func (p JournalerParameters) Validate() error {
	if p.MaxRecords < 1 {
		return fmt.Errorf("max_records must be >= 1")
	}
	if p.BatchFn == nil {
		return fmt.Errorf("batch handler cannot be nil")
	}
	return nil
}

type JournalerMetadata struct {
	StreamUUID string
	Source     string
}

type Journaler struct {
	doneCh        chan struct{}
	inputCh       chan JournalEntry
	batchFn       BatchHandler
	maxRecords    int
	flushInterval time.Duration
}

func NewJournaler(params JournalerParameters) (*Journaler, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	tj := &Journaler{
		doneCh:        make(chan struct{}),
		inputCh:       make(chan JournalEntry, 1<<10),
		maxRecords:    params.MaxRecords,
		flushInterval: params.FlushInterval,
		batchFn:       params.BatchFn,
	}
	go tj.process()
	return tj, nil
}

func (t *Journaler) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	entry := JournalEntry{
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

func (t *Journaler) Log(msg string, metadata JournalerMetadata) {
	entry := JournalEntry{
		Timestamp:  time.Now(),
		Message:    msg,
		StreamUUID: metadata.StreamUUID,
		Source:     metadata.Source,
	}
	select {
	case t.inputCh <- entry:
	default:
	}
}

func (t *Journaler) process() {
	defer close(t.doneCh)

	ticker := time.NewTicker(t.flushInterval)
	defer ticker.Stop()

	buffer := make([]JournalEntry, 0, t.maxRecords)

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

func (t *Journaler) Close() error {
	close(t.inputCh)
	<-t.doneCh
	return nil
}

// TaggerJournaler wraps a Journaler and injects a specific source tag.
type TaggerJournaler struct {
	logger   *Journaler
	metadata JournalerMetadata
}

func NewTaggerJournaler(logger *Journaler, metadata JournalerMetadata) *TaggerJournaler {
	return &TaggerJournaler{
		logger:   logger,
		metadata: metadata,
	}
}

// Write satisfies io.Writer. It delegates to the logger with the pre-defined source.
func (tw *TaggerJournaler) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	tw.logger.Log(string(p), tw.metadata)
	return len(p), nil
}
