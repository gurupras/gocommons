package gocommons

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

type SortInterface interface {
	String() string
	Less(s SortInterface) (bool, error)
}

type SortCollection []SortInterface

type SortParams struct {
	LineConvert func(string) SortInterface
	Lines       SortCollection
}

func (s SortCollection) Len() int {
	return len(s)
}

func (s SortCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortCollection) Less(i, j int) bool {
	if s[i] != nil {
		if ret, err := s[i].Less(s[j]); err == nil {
			return ret
		} else {
			// We can't really send this up. So panic
			panic(fmt.Sprintf("Failed to sort:%v", err))
		}
	} else {
		return false
	}
}

func ExternalSort(file string, bufsize int, sort_params SortParams) (chunks []string, err error) {
	var fstruct *File

	var outfile_path string
	var outfile_raw *File
	var outfile Writer

	chunk_idx := 0
	bytes_read := 0

	//fmt.Printf("Splitting '%s' into chunks\n", file)

	if fstruct, err = Open(file, os.O_RDONLY, GZ_UNKNOWN); err != nil {
		return
	}
	defer fstruct.Close()

	inputChannel := make(chan string, 10000)
	idx := 0
	lines := 0
	go fstruct.AsyncRead(bufio.ScanLines, inputChannel)
	for {
		sort_params.Lines = sort_params.Lines[:0]
		bytes_read = 0
		for {
			line, ok := <-inputChannel
			if !ok {
				break
			}

			if object := sort_params.LineConvert(line); object != nil {
				sort_params.Lines = append(sort_params.Lines, object)
				bytes_read += len(line)
			} else {
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Failed to parse line(%d): \n%s\n", line, idx))
			}
			lines++
			if lines%1000 == 0 {
				//fmt.Println("Lines:", lines)
			}
			if bytes_read > bufsize {
				break
			}
		}
		if len(sort_params.Lines) == 0 {
			// We got no lines in the last iteration..break
			break
		}

		sort.Sort(sort_params.Lines)

		outfile_path = fmt.Sprintf("%s.chunk.%08d.gz", file, chunk_idx)
		//fmt.Println("Saving to chunk:", outfile_path)
		if outfile_raw, err = Open(outfile_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, GZ_TRUE); err != nil {
			return
		}
		defer outfile_raw.Close()

		if outfile, err = outfile_raw.Writer(0); err != nil {
			return
		}
		chunks = append(chunks, outfile_path)

		for idx, object := range sort_params.Lines {
			outfile.Write([]byte(object.String()))
			if idx < len(sort_params.Lines)-1 {
				outfile.Write([]byte("\n"))
			}
		}
		chunk_idx += 1
		outfile.Flush()
		outfile.Close()
	}
	//fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %d lines", file, lines))
	return
}

func NWayMergeGenerator(chunks []string, sort_params SortParams, out_channel chan SortInterface,
	callback func(out_channel chan SortInterface, quit chan bool)) error {

	var readers map[string]*bufio.Scanner = make(map[string]*bufio.Scanner)
	var channels map[string]chan string = make(map[string]chan string)
	var err error
	quit := make(chan bool, 1)

	// Read file and write to channel
	closed_channels := 0
	producer := func(idx int) {
		chunk := chunks[idx]
		reader := readers[chunk]
		channel := channels[chunk]

		reader.Split(bufio.ScanLines)
		lines := 0
		for reader.Scan() {
			line := reader.Text()
			lines++
			channel <- line
			//fmt.Println("CHANNEL-%d: %s", idx, line)
		}
		//fmt.Println("Closing channel:", idx, ":", lines)
		closed_channels++
		close(channel)
	}

	// Now for the consumer
	consumer := func() {
		defer close(out_channel)

		lines := make([]string, len(chunks))
		loglines := make([]SortInterface, len(chunks))

		lines_read := 0
		for {
			var more bool = false
			for idx, chunk := range chunks {
				channel := channels[chunk]
				if loglines[idx] == nil {
					line, ok := <-channel
					lines[idx] = line
					if !ok {
						continue
					}
					lines_read++
					if line != "" {
						if logline := sort_params.LineConvert(line); logline != nil {
							loglines[idx] = logline
							more = true
						} else {
							fmt.Println(fmt.Sprintf("Failed to parse:\n%s\n", line))
						}
					} else {
						fmt.Fprintln(os.Stderr, "Received empty line from channel")
					}
				} else {
					// This index was not nil. This implies we have more
					more = true
				}
			}
			if more == false {
				break
			}

			min := loglines[0]
			min_idx := 0
			for idx, logline := range loglines {
				if logline != nil {
					if less, err := logline.Less(min); err != nil {
						fmt.Fprintln(os.Stderr, fmt.Sprintf("Failed to compare lines: \n%v\n%v\n", logline.String(), min.String()))
						os.Exit(-1)
					} else {
						if less {
							min = logline
							min_idx = idx
						}
					}
				}
			}
			next_line := loglines[min_idx]
			lines[min_idx] = ""
			loglines[min_idx] = nil

			if next_line == nil {
				fmt.Fprintln(os.Stderr, "Attempting to write nil to file..The loop should've broken before this point")
				fmt.Fprintln(os.Stderr, "Open channels:", (len(chunks) - closed_channels))
				fmt.Fprintln(os.Stderr, "lines read:", lines_read)
				for idx, l := range loglines {
					if l != nil {
						fmt.Fprintln(os.Stderr, idx, ":", l.String())
					}
				}
				close(out_channel)
			}
			out_channel <- next_line
		}
	}

	// Set up readers and channels
	for _, chunk := range chunks {
		chunk_file, err := Open(chunk, os.O_RDONLY, GZ_TRUE)
		if err != nil {
			goto out
		}
		defer chunk_file.Close()
		reader, err := chunk_file.Reader(1048576)
		if err != nil {
			goto out
		}
		readers[chunk] = reader
		// Resize channel size based on number of channels
		NORMAL_SIZE := 10000
		channel_size := NORMAL_SIZE / len(chunks)
		channels[chunk] = make(chan string, channel_size)
	}

	// Start the producers
	for idx, _ := range chunks {
		go producer(idx)
	}

	// Start consumer
	go consumer()

	go callback(out_channel, quit)
	//fmt.Println("NWayMergeGenerator: Waiting for callback to quit")
	_ = <-quit
out:
	return err
}
