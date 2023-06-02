package lib

import (
	"strings"
)

type MessageHeader struct {
	header map[string]string
}

func (h *MessageHeader) Has(key string) bool {
	_, ok := h.header[key]
	return ok
}

func (h *MessageHeader) Get(key string) string {
	return h.header[key]
}

func (h *MessageHeader) Set(data string) {
	h.header = map[string]string{}

	arr := strings.Split(data, "\r\n")
	for _, value := range arr[1:] {
		index := strings.Index(value, ":")
		if index >= 0 {
			h.header[value[:index]] = strings.TrimSpace(value[index+1:])
		}
	}
}
