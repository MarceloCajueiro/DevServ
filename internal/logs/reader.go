package logs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"
)

// ReaderOptions configures log reading behavior.
type ReaderOptions struct {
	Follow  bool          // Tail the file for new entries
	Lines   int           // Last N lines (0 = all)
	Since   time.Time     // Filter entries since this time
	Until   time.Time     // Filter entries until this time
	Pattern string        // Regex pattern to filter messages
	Stream  string        // Filter by stream (stdout/stderr)
}

// Reader reads log entries from a file.
type Reader struct {
	file    *os.File
	path    string
	opts    ReaderOptions
	pattern *regexp.Regexp
}

// NewReader creates a new log reader.
func NewReader(path string, opts ReaderOptions) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	r := &Reader{
		file: file,
		path: path,
		opts: opts,
	}

	if opts.Pattern != "" {
		pattern, err := regexp.Compile(opts.Pattern)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		r.pattern = pattern
	}

	return r, nil
}

// ReadAll reads all matching entries from the log file.
func (r *Reader) ReadAll() ([]Entry, error) {
	var entries []Entry

	scanner := bufio.NewScanner(r.file)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip invalid lines
		}

		if r.matches(entry) {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Apply lines limit (return last N)
	if r.opts.Lines > 0 && len(entries) > r.opts.Lines {
		entries = entries[len(entries)-r.opts.Lines:]
	}

	return entries, nil
}

// Read returns a channel of log entries.
// If opts.Follow is true, it will continue tailing the file.
func (r *Reader) Read(ctx context.Context) <-chan Entry {
	ch := make(chan Entry, 100)

	go func() {
		defer close(ch)
		defer r.file.Close()

		// First, read existing entries
		entries, err := r.readExisting()
		if err != nil {
			return
		}

		// Apply lines limit
		if r.opts.Lines > 0 && len(entries) > r.opts.Lines {
			entries = entries[len(entries)-r.opts.Lines:]
		}

		// Send existing entries
		for _, entry := range entries {
			select {
			case ch <- entry:
			case <-ctx.Done():
				return
			}
		}

		// If not following, we're done
		if !r.opts.Follow {
			return
		}

		// Tail the file for new entries
		r.tail(ctx, ch)
	}()

	return ch
}

func (r *Reader) readExisting() ([]Entry, error) {
	var entries []Entry

	scanner := bufio.NewScanner(r.file)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if r.matches(entry) {
			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

func (r *Reader) tail(ctx context.Context, ch chan<- Entry) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	reader := bufio.NewReader(r.file)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					return
				}

				var entry Entry
				if err := json.Unmarshal(line, &entry); err != nil {
					continue
				}

				if r.matches(entry) {
					select {
					case ch <- entry:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

func (r *Reader) matches(entry Entry) bool {
	// Filter by time range
	if !r.opts.Since.IsZero() && entry.Timestamp.Before(r.opts.Since) {
		return false
	}
	if !r.opts.Until.IsZero() && entry.Timestamp.After(r.opts.Until) {
		return false
	}

	// Filter by stream
	if r.opts.Stream != "" && entry.Stream != r.opts.Stream {
		return false
	}

	// Filter by pattern
	if r.pattern != nil && !r.pattern.MatchString(entry.Message) {
		return false
	}

	return true
}

// Close closes the reader.
func (r *Reader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}
