package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"read2succeed/data"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/unknwon/paginater"
)

// GetGroupReadings - list user's/users' read books
func (s *Service) GetGroupReadings(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*data.AuthUser)

	tmplData := struct {
		GroupID      int
		Readings     []data.GroupReading
		Page         *paginater.Paginater
		PageQueryRaw string
		GroupList    []data.GroupMemberCount
	}{}

	groupIDStr := r.URL.Query().Get("group")
	if groupIDStr == "" {
		userGroups, err := s.store.GetUserGroupMemberCount(user.ID)
		if err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		tmplData.GroupList = userGroups
		//fmt.Printf("tmplData.GroupList: %+v", tmplData.GroupList)
	} else {
		groupID, err := strconv.Atoi(groupIDStr)
		if err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		var page int
		itemsPerPage := 20
		pageParam := r.URL.Query().Get("page")
		page, err = strconv.Atoi(pageParam)
		if err != nil {
			page = 1
		}

		offset := (page - 1) * itemsPerPage
		req := data.GroupReadingsRequest{
			UserID:  user.ID,
			GroupID: groupID,
			Limit:   &itemsPerPage,
			Offset:  &offset,
		}
		output, err := s.store.ListUserGroupsReadings(req)
		if err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmplData.GroupID = groupID
		tmplData.Page, tmplData.PageQueryRaw = _buildPaginater(output.Pagination.Total, itemsPerPage, page, r)

		tmplData.Readings = output.GroupReadings
		//tmplData.Page = p
		//tmplData.PageQueryRaw = qqs

	}

	//s.l.Printf("stats: %#v\n", stats)

	if err := s.t.ExecuteTemplate(rw, "group_readings.gohtml", tmplData); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// AddGroup - add new group
func (s *Service) AddGroup(rw http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(rw, "Invalid request: expecting JSON.", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	newGroup := &data.Group{}
	err := decoder.Decode(newGroup)
	if err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	user := r.Context().Value("user").(*data.AuthUser)
	newGroup.UserID = user.ID
	log.Println(newGroup)
	err = s.store.AddGroup(newGroup)
	if err != nil {
		s.l.Printf("AddGroup(%d, %s):", user.ID, newGroup.Name)
		s.l.Println(err)
		http.Error(rw, "{\"status\":\"error\", \"message\":\"Unable to add group\"}", http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("{\"status\":\"ok\"}"))
}

// UpdateGroup - update the given group
func (s *Service) UpdateGroup(rw http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("user").(*data.AuthUser)
	vars := mux.Vars(r)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	groupID, _ := strconv.Atoi(vars["id"])
	group, err := s.store.GetGroupByID(groupID)
	if err != nil {
		s.l.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	//fmt.Printf("groupbyid: %v\n", group)
	//fmt.Println("about to compare", user.ID, "with", group.UserID)
	if group.UserID != user.ID {
		err := errors.New("not allowed")
		http.Error(rw, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	updGroup := data.UpdateGroup{}
	if err := decoder.Decode(&updGroup); err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}
	//s.l.Printf("updGroup: %+v", updGroup)
	if updGroup.Name == "" {
		updGroup.Name = group.Name
	}
	if updGroup.Status == "" {
		updGroup.Status = group.Status
	}

	if err = s.store.UpdateGroup(group.ID, updGroup); err != nil {
		s.l.Println("UpdateGroup:", err)
		http.Error(rw, "{\"status\":\"error\", \"message\":\"Unable to update group\"}", http.StatusBadRequest)
		return
	}

	rw.Write([]byte("{\"status\":\"ok\"}"))
}

// FindAvailableGroups - find group a reader can join to
func (s *Service) FindAvailableGroups(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Cache-Control", "no-cache")
	var err error

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	params := struct {
		Query    string `json:"name"`
		ReaderID int    `json:"reader"`
	}{}
	if err = decoder.Decode(&params); err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	groups := []data.Group{}
	if len(params.Query) < 3 {
		encoder := json.NewEncoder(rw)
		encoder.Encode(groups)
		return
	}

	fmt.Printf("FindNewGroupsForReader > params: %+v\n", params)
	// TODO - also pass the current user_id
	groups, err = s.store.FindNewGroupsForReader(params.Query, params.ReaderID)
	if err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusInternalServerError)
		return
	}
	//fmt.Printf("groups: %+v\n", groups)
	encoder := json.NewEncoder(rw)
	encoder.Encode(groups)
}

// JoinGroup - join a group
func (s *Service) JoinGroup(rw http.ResponseWriter, r *http.Request) {

	// TODO - this needs some work as anyone can join any group :(

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Cache-Control", "no-cache")

	//user := r.Context().Value("user").(*data.AuthUser)
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	params := struct {
		GroupID  int    `json:"group"`
		ReaderID int    `json:"reader"`
		Query    string `json:"query"`
	}{}
	if err := decoder.Decode(&params); err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	fmt.Printf("join a group params: %+v\n", params)

	group, err := s.store.GetGroupByID(params.GroupID)
	if err != nil {
		http.Error(rw, `{"status":"error"}`, http.StatusBadRequest)
		return
	}

	if group.Status == "private" && group.AccessCode != params.Query {
		s.l.Printf("can't join reader %d to private group %d\n", params.ReaderID, group.ID)
		http.Error(rw, `{"status":"error"}`, http.StatusBadRequest)
		return
	}

	reader, err := s.store.GetReaderByID(params.ReaderID)
	if err != nil {
		s.l.Println(err.Error())
		http.Error(rw, `{"status":"error", "message":"Reader not found!"}`, http.StatusInternalServerError)
		return
	}

	user := r.Context().Value("user").(*data.AuthUser)

	// making sure it's user's reader
	if reader.UserID != user.ID {
		http.Error(rw, `{"status":"error", "message":"Not allowed!"}`, http.StatusBadRequest)
		return
	}

	err = s.store.GroupAddReader(group.ID, reader.ID)
	if err != nil {
		rw.Write([]byte(`{"status":"error", "message":"Cannot join group!"}`))
		return
	}

	rw.Write([]byte("{\"status\":\"ok\"}"))
}

// LeaveGroup -Leave a group
func (s *Service) LeaveGroup(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Cache-Control", "no-cache")

	user := r.Context().Value("user").(*data.AuthUser)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	params := struct {
		GroupID  int `json:"group_id"`
		ReaderID int `json:"reader_id"`
	}{}
	if err := decoder.Decode(&params); err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	fmt.Printf("params: %+v\n", params)

	group, err := s.store.GetGroupByID(params.GroupID)
	if err != nil {
		http.Error(rw, `{"status":"error"}`, http.StatusBadRequest)
		return
	}
	fmt.Printf("group: %+v\n", group)

	reader, err := s.store.GetReaderByID(params.ReaderID)
	if err != nil {
		s.l.Println(err.Error())
		http.Error(rw, `{"status":"error", "message":"Reader not found!"}`, http.StatusInternalServerError)
		return
	}

	if reader.UserID != user.ID {
		http.Error(rw, `{"status":"error", "message":"Not allowed!"}`, http.StatusBadRequest)
		return
	}

	// leave the group
	err = s.store.GroupRemoveReader(params.GroupID, params.ReaderID)
	if err != nil {
		log.Println("Unable to remove the reader from the group: ", err)
		http.Error(rw, `{"status":"error", "message":"Unable to remove the reader from the group!"}`,
			http.StatusInternalServerError)
		return
	}
	rw.Write([]byte(`{"status":"ok"}`))
}

// GetGroupReaders - list user in a group (only admins can access this part)
func (s *Service) GetGroupReaders(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*data.AuthUser)
	vars := mux.Vars(r)

	groupIDStr := vars["id"]
	if !user.IsAdmin {
		s.l.Printf("user %d/%s not allowed on /groupreaders?group=%s", user.ID, user.Email, groupIDStr)
		http.Error(rw, `Page Not Found`, http.StatusNotFound)
		return
	}

	tmplData := struct {
		Group        data.Group
		GroupReaders []data.Reader
	}{}

	if groupIDStr != "" {
		groupID, err := strconv.Atoi(groupIDStr)
		if err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		group, err := s.store.GetGroupByID(groupID)
		if err != nil {
			s.l.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		readers, err := s.store.GetGroupReaders(groupID)
		if err != nil {
			s.l.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmplData.Group = group
		tmplData.GroupReaders = readers
	}

	//s.l.Printf("stats: %#v\n", tmplData)

	if err := s.t.ExecuteTemplate(rw, "group_readers.gohtml", tmplData); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

}
