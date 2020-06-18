package routing

import (
	"bytes"
	"mobchat/util"
)

//Route -
type Route struct {
	Path   []*Node
	ignore bool
}

func filterIgnore(routes []Route) []Route {
	retAr := make([]Route, 0)
	for _, route := range routes {
		if !route.ignore {
			retAr = append(retAr, route)
		}
	}
	return retAr
}

func routes(paths [][][]byte) []Route {
	routes := make([]Route, len(paths))
	for i, path := range paths {
		routes[i] = Route{}
		for _, id := range path {
			mutex.Lock()
			node, exists := Table.Nodes[util.ToHexString(id)]
			if !exists {
				routes[i].ignore = true
				mutex.Unlock()
				break
			}
			routes[i].Path = append(routes[i].Path, node)
			mutex.Unlock()
		}
	}
	return filterIgnore(routes)
}

func findShortest(paths [][][]byte) [][][]byte {
	shortest := int(99)
	for _, path := range paths {
		if len(path) < shortest {
			shortest = len(path)
		}
	}
	shortestPaths := make([][][]byte, 0)
	for _, path := range paths {
		if path == nil {
			continue
		}
		if len(path) == shortest {
			shortestPaths = append(shortestPaths, path)
		}
	}
	if len(shortestPaths) == 0 {
		return nil
	}
	return shortestPaths
}

func find(startPath [][]byte, ID []byte) [][][]byte {
	if len(startPath) > 10 {
		return nil
	}
	paths := make([][][]byte, 0)
	startNode := Table.Nodes[util.ToHexString(startPath[len(startPath)])]
	for key := range startNode.Connections {
		n := startNode.Connections[key]
		path := append(startPath, n.ID())
		if bytes.Equal(ID, n.ID()) {
			paths = append(paths, path)
		} else {
			paths = append(paths, find(path, ID)...)
		}
	}
	return findShortest(paths)
}

//FindRoute -
func (routing *Routing) FindRoute(findID []byte, startIDs [][]byte) []Route {
	mutex.Lock()
	shortestPaths := make([][][]byte, 0)
	for _, id := range startIDs {
		shortestPaths = append(shortestPaths, find([][]byte{id}, findID)...)
	}
	mutex.Unlock()
	shortestPaths = findShortest(shortestPaths)
	return routes(shortestPaths)
}

//SerializeRoutes -
func SerializeRoutes(routes []Route) []byte {
	byteAr := make([][]byte, len(routes))
	for i, route := range routes {
		var buff bytes.Buffer
		buff.WriteByte(byte(len(route.Path)))
		for _, node := range route.Path {
			buff.Write(node.Serialize())
		}
		byteAr[i] = buff.Bytes()
	}
	return bytes.Join(byteAr, []byte{})
}

//DeserializeRoutes -
func DeserializeRoutes(data []byte) ([]Route, error) {
	idx := 0
	routes := make([]Route, 0)
	for idx < len(data) {
		ln := int(data[idx])
		idx++
		route := Route{}
		for idx < idx+144*ln {
			node, err := DeserializeNode(data[idx : idx+144])
			if err != nil {
				return nil, err
			}
			route.Path = append(route.Path, &node)
			idx += 144
		}
		routes = append(routes, route)
	}
	return routes, nil
}
