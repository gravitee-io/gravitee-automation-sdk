# AM mock server

In-memory HTTP mock of the AM Automation API. Flags, auth and dry-run behaviour are described in the
[root README](../README.md#mock-server).

## Fault injection

A running mock can answer an error instead of serving a request, for example to refuse a delete and then allow it, without losing what it stores. A fault matches a method (any case) and the exact request path, base path included (the query is ignored), and answers `status` (400-599) with AM's error body `{"http_status": <status>, "message": <message>}`. `times` > 0 answers that many requests, then the fault clears itself; omitted, it stays until cleared. A fault answers after authentication, authorization and the domain parent check, and before spec validation and dry-run.

In Go, through `MockAM.Faults`:

```go
mockAM.Faults.Set(server.Fault{Method: "DELETE", Path: path, Status: 400, Message: "You can't delete a certificate with existing applications."})
mockAM.Faults.Clear("DELETE", path)
mockAM.Faults.Reset()
```

Over HTTP, at `/_mock/faults` (always on, outside the base path, no token needed):

```bash
# Refuse a certificate delete
curl -X PUT localhost:8080/_mock/faults -H 'Content-Type: application/json' -d '{
  "method": "DELETE",
  "path": "/automation/organizations/DEFAULT/environments/DEFAULT/domains/my-domain/certificates/my-cert",
  "status": 400,
  "message": "You can'"'"'t delete a certificate with existing applications."
}'
# List the faults in place
curl localhost:8080/_mock/faults
# Clear that fault, then every fault
curl -X DELETE 'localhost:8080/_mock/faults?method=DELETE&path=/automation/organizations/DEFAULT/environments/DEFAULT/domains/my-domain/certificates/my-cert'
curl -X DELETE localhost:8080/_mock/faults
```

`PUT` answers `204`, or `400` for an invalid fault. `DELETE` answers `204`, or `400` when only one of `method` and `path` is given.
