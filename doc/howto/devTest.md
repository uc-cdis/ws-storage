# TL;DR

ws-storage requires `golang 1.14+`.  It implements a golang module.


## Build and Test

See the [Dockerfile](../../Dockerfile):

Update dependencies with:
```
go get -u
```

Setup AWS creds:
```
export AWS_REGION=us-east-1
export AWS_STS_REGIONAL_ENDPOINTS=regional
```
and
```
eval "$(ssh reuben@cdistest.csoc 'source cloud-automation/gen3/gen3setup.sh; gen3 arun env' | grep AWS_ | grep -v AWS_PROFILE | awk '{ print export $0 }')"
```

Build and test:
```
go build -o bin/ws-storage
go test -v ./storage/ -failfast
```

```
gen3 arun env | grep AWS | awk '{ print "export " $0 }'
```
