# Overview

This services provides methods for managing `objects` in `workspaces`.
Each object has a unique `key` within a particular `workspace`.
Each user has full access to a personal workspace.
In the future we may extend the system to allow a user to
access other shared workspaces.

## API

```
GET|DELETE /ws-storage/list/workspace/key
GET /ws-storage/upload/workspace/key
GET /ws-storage/download/workspace/key
```

The `REMOTE_USER` header is set at the api gateway (revproxy) after verifying the access token's authentication and authorization.  A user with the `workspace` role is authorized to access workspace storage.

Currently only the `@user` workspace is supported - which corresponds to the user's personal storage space.

## Implementation

This service is just a thin wrapper around the S3 API providing controlled access to an S3 bucket.  A work space is just a particular
prefix in the backing S3 bucket.  This implementation should integrte directly with mariner.


## References

* original [design doc](https://docs.google.com/document/d/1UK59Dplttu4KKIoymZwwmuOpmU6kmdu_FZRwCP-4vCM/edit#heading=h.ra5thdn8e3xr)

