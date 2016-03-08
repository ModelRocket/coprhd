package coprhd

import (
	"fmt"
	"time"
)

const (
	CreateExportUri    = "block/exports.json"
	QueryExportUriTpl  = "block/exports/%s.json"
	SearchExportUri    = "block/exports/search.json?"
	DeleteExportUriTpl = "block/exports/%s/deactivate.json"

	ExportTypeExclusive = "Exclusive"
)

type (
	ExportService struct {
		*Client
		id         string
		itrs       []string
		project    string
		exportType ExportType
		array      string
		volumes    []ResourceId
	}

	Export struct {
		BaseObject    `json:",inline"`
		Volumes       []ResourceId    `json:"volumes"`
		Initiators    []Initiator     `json:"initiators"`
		Hosts         []NamedResource `json:"hosts"`
		Clustsers     []NamedResource `json:"clusters"`
		Type          string          `json:"type"`
		GeneratedName string          `json:"generated_name"`
		PathParams    []string        `json:"path_params"`
	}

	CreateExportReq struct {
		Initiators []string     `json:"initiators"`
		Name       string       `json:"name"`
		Project    string       `json:"project"`
		Type       ExportType   `json:"type"`
		VArray     string       `json:"varray"`
		Volumes    []ResourceId `json:"volumes"`
	}

	ExportType string
)

// Export gets an instance to the ExportService
func (this *Client) Export() *ExportService {
	return &ExportService{
		Client:     this.Copy(),
		itrs:       make([]string, 0),
		volumes:    make([]ResourceId, 0),
		exportType: ExportTypeExclusive,
	}
}

func (this *ExportService) Id(id string) *ExportService {
	this.id = id
	return this
}

func (this *ExportService) Initiators(itrs ...string) *ExportService {
	this.itrs = append(this.itrs, itrs...)
	return this
}

func (this *ExportService) Volumes(vols ...string) *ExportService {
	for _, v := range vols {
		this.volumes = append(this.volumes, ResourceId{v})
	}
	return this
}

func (this *ExportService) Project(project string) *ExportService {
	this.project = project
	return this
}

func (this *ExportService) Array(array string) *ExportService {
	this.array = array
	return this
}

func (this *ExportService) Type(t ExportType) *ExportService {
	this.exportType = t
	return this
}

// Create creates and export with the specfied name
func (this *ExportService) Create(name string) (*Export, error) {
	req := CreateExportReq{
		Name:       name,
		Initiators: this.itrs,
		Project:    this.project,
		Type:       this.exportType,
		VArray:     this.array,
		Volumes:    this.volumes,
	}

	task := Task{}

	err := this.Post(CreateExportUri, &req, &task)
	if err != nil {
		if this.LastError().IsExportVolDup() {
			return this.Search("name=" + name)
		}
		return nil, err
	}

	// wait for the task to complete
	err = this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)

	if err != nil {
		return nil, err
	}

	this.id = task.Resource.Id

	return this.Query()
}

func (this *ExportService) Query() (*Export, error) {
	path := fmt.Sprintf(QueryExportUriTpl, this.id)
	exp := Export{}

	err := this.Get(path, nil, &exp)
	if err != nil {
		return nil, err
	}

	return &exp, nil
}

func (this *ExportService) Search(query string) (*Export, error) {
	path := SearchExportUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}

func (this *ExportService) Delete(id string) error {
	path := fmt.Sprintf(DeleteExportUriTpl, id)

	task := Task{}

	err := this.Post(path, nil, &task)
	if err != nil {
		return err
	}

	return this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
}
