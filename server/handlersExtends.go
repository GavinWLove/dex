package server

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)


// listConnectors list connector
func (s *Server) listConnectors(w http.ResponseWriter, r *http.Request){
	s.logger.Infof("listConnector")


	connectors, err := s.storage.ListConnectors()
	if err != nil {
		s.logger.Errorf("Failed to get list of connectors: %v", err)
		//s.renderError(r, w, http.StatusInternalServerError, "Failed to retrieve connector list.")
		return
	}
	// Construct a URL with all of the arguments in its query
	connURL := url.URL{
		RawQuery: r.Form.Encode(),
	}
	d := connInfos{
		Path: "test",
		Data: make([]connectorInfo, len(connectors)),
	}
	for index, conn := range connectors {
		connURL.Path = s.absPath("/auth", conn.ID)
		d.Data[index] = connectorInfo{
			ID:   conn.ID,
			Name: conn.Name,
			Type: conn.Type,
			URL:  connURL.String(),
		}
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)

}
type connInfos struct {
	Path	string
	Data  []connectorInfo `json:"data"`
}
