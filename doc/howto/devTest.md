# TL;DR

ws-storage requires `golang 1.14+`.  It implements a golang module.

## Test Credentials

Setup AWS creds:
```
export AWS_REGION=us-east-1
export AWS_STS_REGIONAL_ENDPOINTS=regional

eval "$(ssh reuben@cdistest.csoc 'source cloud-automation/gen3/gen3setup.sh; gen3 arun env' | grep AWS_ | grep -v AWS_PROFILE | awk '{ print "export " $0 }')"
```

## Test Setup

```
(
    set -e
    bucket="$(jq -r .bucket < testData/testConfig.json)"
    prefix="$(jq -r .bucketprefix < testData/testConfig.json)"
    for name in x y z subfolder1/x; do
        echo "some randome stuff" | aws s3 cp - "s3://${bucket}/${prefix}/goTestUser/goTestSuite/$name"
    done
)
```

## Build and Test

See the [Dockerfile](../../Dockerfile):

Update dependencies with:
```
go get -u
```

Build and test:
```
go build -o bin/ws-storage
go test -v ./storage/ -failfast
```

```
gen3 arun env | grep AWS | awk '{ print "export " $0 }'
```

# Interacting with the server

```
curl http://localhost:8000/ws-storage/list/@user/
curl -H 'REMOTE_USER: goTestUser' http://localhost:8000/ws-storage/list/@user/
curl -H 'REMOTE_USER: goTestUser' http://localhost:8000/ws-storage/list/@user/goTestSuite/ | jq -r .
curl -H 'REMOTE_USER: goTestUser' http://localhost:8000/ws-storage/download/@user/goTestSuite/z | jq -r .
```

Note that the server relies on the API gateway (revproxy) for authentication and authorization.
