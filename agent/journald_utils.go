package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

//  apt-get install libsystemd-dev

func WatchJournalLogs(key string, value string, fields []string, sinceTime time.Time, dataChannel chan map[string]string, stopChannel chan struct{}) error {
	journal, err := sdjournal.NewJournal()
	if err != nil {
		return fmt.Errorf("failed to open journal: %w", err)
	}
	defer journal.Close()

	// Add match filter
	m := sdjournal.Match{Field: key, Value: value}
	if err := journal.AddMatch(m.String()); err != nil {
		return fmt.Errorf("failed to add journal match: %w", err)
	}

	// Process past logs if sinceTime is specified
	if !sinceTime.IsZero() {
		// Seek to the specified timestamp
		usec := uint64(sinceTime.UnixMicro())
		if err := journal.SeekRealtimeUsec(usec); err != nil {
			return fmt.Errorf("failed to seek to timestamp: %w", err)
		}

		// Read all existing entries
		for {
			n, err := journal.Next()
			if err != nil {
				return fmt.Errorf("error reading next journal entry: %w", err)
			}

			if n == 0 {
				break // No more entries
			}

			data := make(map[string]string)
			for _, field := range fields {
				val, err := journal.GetData(field)
				if err == nil {
					data[field] = val
				}
			}

			// Add timestamp information
			ts, err := journal.GetRealtimeUsec()
			if err == nil {
				data["__REALTIME_TIMESTAMP"] = strconv.FormatUint(ts, 10)
			}

			select {
			case dataChannel <- data:
			case <-stopChannel:
				return nil
			}
		}
	}

	// Now start following new logs
	if err := journal.SeekTail(); err != nil {
		return fmt.Errorf("failed to seek to journal tail: %w", err)
	}
	journal.Next() // Advance past the last entry

	for {
		select {
		case <-stopChannel:
			return nil
		default:
			ret := journal.Wait(sdjournal.IndefiniteWait)
			if ret == sdjournal.SD_JOURNAL_NOP {
				continue
			}

			for {
				n, err := journal.Next()
				if err != nil {
					return fmt.Errorf("error reading next journal entry: %w", err)
				}

				if n == 0 {
					break
				}

				data := make(map[string]string)
				for _, field := range fields {
					val, err := journal.GetData(field)
					if err == nil {
						data[field] = val
					}
				}

				// Add timestamp information
				ts, err := journal.GetRealtimeUsec()
				if err == nil {
					data["__REALTIME_TIMESTAMP"] = strconv.FormatUint(ts, 10)
				}

				select {
				case dataChannel <- data:
				case <-stopChannel:
					return nil
				}
			}
		}
	}
}
