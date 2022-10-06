package jsondb

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// DB structure that represent a JSON database.
type DB[T any] struct {
	FilePath string
	fd       *os.File

	mtx *sync.Mutex

	currentData []T

	toggleRead  chan struct{}
	toggleWrite chan T

	errChan chan error
}

// Connect constructor for DB struct.
func Connect[T any](file string) (*DB[T], error) {
	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}
	return &DB[T]{
		FilePath:    file,
		fd:          fd,
		mtx:         new(sync.Mutex),
		toggleRead:  make(chan struct{}),
		toggleWrite: make(chan T),
		errChan:     make(chan error),
	}, nil
}

func (j *DB[T]) readFromFile() error {
	if len(j.currentData) != 0 {
		j.fd.Seek(0, 0)
	}
	data, err := io.ReadAll(j.fd)
	if err != nil {
		return err
	}

	var d []T

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	j.currentData = d
	return nil
}

func (j *DB[T]) startWriteListener() {
	go func() {
		for {
			v, ok := <-j.toggleWrite
			if !ok {
				return
			}

			j.mtx.Lock()
			j.currentData = append(j.currentData, v)

			j.fd.Seek(0, 0)
			j.fd.Truncate(0)
			if err := json.NewEncoder(j.fd).Encode(j.currentData); err != nil {
				j.errChan <- err
				j.mtx.Unlock()
				continue
			}
			j.errChan <- nil
			j.mtx.Unlock()
		}
	}()
}

func (j *DB[T]) startReadListener() {
	go func() {
		for {
			<-j.toggleRead
			j.mtx.Lock()
			err := j.readFromFile()
			if err != nil {
				j.errChan <- err
				j.mtx.Unlock()
				continue
			}
			j.errChan <- nil
			j.mtx.Unlock()
		}
	}()
}

func (j *DB[T]) Init() error {
	if err := j.readFromFile(); err != nil {
		return err
	}
	j.startReadListener()
	j.startWriteListener()
	return nil
}

// Close closes the JSON file db connection.
func (j *DB[T]) Close() error {
	//
	close(j.toggleRead)
	close(j.toggleWrite)
	close(j.errChan)
	//
	return j.fd.Close()
}

// WriteToDB write data into JSON file. returns an error if there is any
// failure when writing data into the JSON file.
func (j DB[T]) WriteToDB(value T) error {
	j.toggleWrite <- value
	if err := <-j.errChan; err != nil {
		return err
	}
	return nil
}

// ReadFromDB read all data that exists in the JSON table. returns an array of
// JSON object or an error.
func (j DB[T]) ReadFromDB() ([]T, error) {
	j.toggleRead <- struct{}{}
	if err := <-j.errChan; err != nil {
		return nil, err
	}
	return j.currentData, nil
}

// UpdateDB updates a value in JSON database by giving the index and the value
// that you want to be the replacement of the old data. function retunrs an
// error, if the index is greater that the length of the JSON object array.
func (j *DB[T]) UpdateDB(index int, v T) error {
	if index >= len(j.currentData) {
		return fmt.Errorf("wrong index")
	}
	j.mtx.Lock()
	defer j.mtx.Unlock()
	j.currentData[index] = v
	j.fd.Seek(0, 0)
	j.fd.Truncate(0)
	return json.NewEncoder(j.fd).Encode(j.currentData)
}

// DeleteFromDB deletes a value from the JSON array data. function accepts the
// index of the data and returns an error if there is any failures.
func (j *DB[T]) DeleteFromDB(index int) error {
	if index >= len(j.currentData) {
		return fmt.Errorf("index is greater than table data length")
	}
	j.mtx.Lock()
	defer j.mtx.Unlock()
	j.currentData = append(j.currentData[:index], j.currentData[index+1:]...)
	j.fd.Seek(0, 0)
	j.fd.Truncate(0)
	return json.NewEncoder(j.fd).Encode(j.currentData)
}
