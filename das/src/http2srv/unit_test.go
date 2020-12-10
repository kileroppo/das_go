package http2srv

import "testing"

var (
	lancensMsg = `
{
  "pushData": {
    "event_guid": "E7B408ED-1915-4060-A05A-5C77E2C55E8E",
    "event_type": "0",
    "uid": "LSV070AAZ8BTF6J54I7W",
    "img": "aHR0cDovL2xhbmNlbnMwaGIub3NzLWNuLXFpbmdkYW8uYWxpeXVuY3MuY29tLzAzQjc3M0QxLURBMDctNEIxQy1CM0EyLTE3QkQyRUEzMjU2NS5qcGc/T1NTQWNjZXNzS2V5SWQ9TFRBSU9pdUFLWUV6THBJOCZFeHBpcmVzPTE2MDcyNTIzMzEmU2lnbmF0dXJlPVIySlRMWWVpRkFJT2czT2ZyM1kwaFZlMmk0dyUzRA==",
    "info": "eyJvcmllbnRhdGlvbiI6IjAifQ=="
  }
}
`
)

func TestProcessLancensMsg(t *testing.T) {
	ProcessLancensMsg(lancensMsg)
}
