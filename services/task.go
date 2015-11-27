package services

import (
	"encoding/json"
	"time"
	"strings"
	"github.com/cloud-ca/go-cloudca/api"
)

//Task status
const (
   PENDING = "PENDING"
   SUCCESS = "SUCCESS"
   FAILED = "FAILED"
)

//A Task object. This object can be used to poll asynchronous operations.
type Task struct {
	Id string
	Status string
	Created string
	Result []byte
}

type TaskService interface {
	Find(id string) (*Task, error)
	Poll(id string, milliseconds time.Duration) ([]byte, error)
}

type TaskApi struct {
 	apiClient api.CcaApiClient
}

func NewTaskService(apiClient api.CcaApiClient) TaskService {
	return &TaskApi{
		apiClient: apiClient,
	}
}

//Retrieve a Task with sepecified id
func (taskApi *TaskApi) Find(id string) (*Task, error) {
	request := api.CcaRequest{
		Method: api.GET,
		Endpoint: "tasks/" + id,
	}
	response, err := taskApi.apiClient.Do(request)
	if err != nil {
		return nil, err
	} else if len(response.Errors) > 0 {
		return nil, api.CcaErrorResponse(*response)
	}
	data := response.Data
	taskMap := map[string]*json.RawMessage{}
	json.Unmarshal(data, &taskMap)
	
	task := Task{}
	json.Unmarshal(*taskMap["id"], &task.Id)
	json.Unmarshal(*taskMap["status"], &task.Status)
	json.Unmarshal(*taskMap["created"], &task.Created)
	if val, ok := taskMap["result"]; ok {
		task.Result = []byte(*val)
	}
	return &task, nil
}

//Poll an the Task API. Blocks until success or failure
func (taskApi *TaskApi) Poll(id string, milliseconds time.Duration) ([]byte, error) {
	ticker := time.NewTicker(time.Millisecond * milliseconds)
	task, err := taskApi.Find(id)
	if err != nil {
		return nil, err
	}
	done := task.Completed()
	for !done {
		<-ticker.C
		task, err = taskApi.Find(id)
		if err != nil {
			return nil, err
		}
		done = task.Completed()
	}
	return task.Result, nil
}

func (task Task) Completed() bool {
   return !strings.EqualFold(task.Status, PENDING)
}